package httpapi

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"aroma-maintenance/internal/application"
)

type Handler struct {
	service *application.MaintenanceService
	page    *template.Template
}

func NewHandler(service *application.MaintenanceService) http.Handler {
	h := &Handler{service: service, page: template.Must(template.New("products").Parse(pageTemplate))}
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.productsPage)
	mux.HandleFunc("/api/products", h.products)
	mux.HandleFunc("/api/products/", h.productAction)
	mux.HandleFunc("/api/logs", h.logs)
	mux.HandleFunc("/api/tasks", h.tasks)
	mux.HandleFunc("/api/tasks/", h.taskAction)
	return mux
}

func (h *Handler) productsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := h.page.Execute(w, map[string]any{"Products": h.service.ListProducts()}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) products(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, h.service.ListProducts())
}

func (h *Handler) productAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "products" {
		http.NotFound(w, r)
		return
	}
	id, action := parts[2], parts[3]
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		Price    string `json:"price"`
		ImageURL string `json:"imageUrl"`
		Delta    int    `json:"delta"`
		Reason   string `json:"reason"`
	}
	if err := decodeJSON(r.Body, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var result any
	var err error
	switch action {
	case "price":
		result, err = h.service.UpdatePrice(id, input.Price)
	case "image":
		result, err = h.service.UploadImage(id, input.ImageURL)
	case "stock":
		result, err = h.service.ChangeStock(id, input.Delta, input.Reason)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) logs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, h.service.Logs())
}

func (h *Handler) tasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, h.service.Tasks())
}

func (h *Handler) taskAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "tasks" || parts[3] != "run" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		Worker string `json:"worker"`
	}
	if err := decodeJSON(r.Body, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if input.Worker == "" {
		input.Worker = "http-worker"
	}
	writeJSON(w, http.StatusOK, h.service.RunTask(parts[2], input.Worker))
}

func decodeJSON(reader io.Reader, destination any) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

const pageTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>香薰产品维护台</title>
<style>
:root{font-family:system-ui,-apple-system,"PingFang SC",sans-serif;color:#2f2925;background:#f7f2eb}body{margin:0}header{padding:28px 7vw;background:#302b29;color:#fff}main{max-width:1100px;margin:32px auto;padding:0 20px}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:18px}.card{background:#fff;border:1px solid #e6ddd2;border-radius:8px;padding:18px;box-shadow:0 4px 14px #2f292512}.tag{color:#96714d;font-size:13px}.price{font-size:24px;font-weight:700;margin:14px 0 8px}.stock{color:#5e655b}button{border:0;border-radius:5px;background:#96714d;color:#fff;padding:9px 14px;cursor:pointer}button:hover{background:#735438}#message{min-height:24px;margin:20px 0;color:#735438}
</style></head><body><header><h1>香薰产品维护台</h1><p>浏览商品，维护价格、图片与库存。</p></header><main><div id="message"></div><section class="grid">{{range .Products}}<article class="card"><div class="tag">{{.Type}}</div><h2>{{.Name}}</h2><div class="price">¥{{.Price}}</div><div class="stock">库存 {{.Stock}}</div><button data-id="{{.ID}}">记录补货 +1</button></article>{{end}}</section></main><script>document.querySelectorAll('button[data-id]').forEach(function(button){button.addEventListener('click',function(){fetch('/api/products/'+button.dataset.id+'/stock',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({delta:1,reason:'前台补货'})}).then(function(r){if(!r.ok)throw new Error('操作失败');return r.json()}).then(function(product){document.getElementById('message').textContent=product.name+' 库存已更新为 '+product.stock}).catch(function(error){document.getElementById('message').textContent=error.message})})})</script></body></html>`
