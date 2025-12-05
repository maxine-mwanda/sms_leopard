package controllers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/example/smsleopard/models"
    "github.com/example/smsleopard/queue"
)

type Handler struct{
    Svc *models.Service
    Pub *queue.Publisher
}

func NewHandler(svc *models.Service, pub *queue.Publisher) *Handler {
    return &Handler{Svc: svc, Pub: pub}
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request){
    var req struct{ Name, Template string }
    if err := json.NewDecoder(r.Body).Decode(&req); err !=nil { w.WriteHeader(http.StatusBadRequest); return }
    if req.Name=="" || req.Template=="" { w.WriteHeader(http.StatusBadRequest); return }
    c, err := h.Svc.CreateCampaign(req.Name, req.Template)
    if err!=nil { w.WriteHeader(http.StatusInternalServerError); return }
    json.NewEncoder(w).Encode(c)
}

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request){
    q := r.URL.Query()
    limit := 10; offset := 0
    if l := q.Get("limit"); l!="" { if v, err := strconv.Atoi(l); err==nil { limit = v } }
    if o := q.Get("offset"); o!="" { if v, err := strconv.Atoi(o); err==nil { offset = v } }
    cs, err := h.Svc.ListCampaigns(limit, offset)
    if err!=nil { w.WriteHeader(http.StatusInternalServerError); return }
    json.NewEncoder(w).Encode(cs)
}

func (h *Handler) SendCampaign(w http.ResponseWriter, r *http.Request){
    var req struct{ CampaignID int64 `json:"campaign_id"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err!=nil{ w.WriteHeader(http.StatusBadRequest); return }
    if req.CampaignID==0 { w.WriteHeader(http.StatusBadRequest); return }
    if err := h.Svc.EnqueueCampaign(req.CampaignID); err!=nil { w.WriteHeader(http.StatusInternalServerError); return }
    if err := h.Pub.PublishSend(req.CampaignID); err!=nil { w.WriteHeader(http.StatusInternalServerError); return }
    json.NewEncoder(w).Encode(map[string]string{"status":"accepted"})
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request){
    var req struct{
        Template string `json:"template"`
        Customer map[string]string `json:"customer"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err!=nil{ w.WriteHeader(http.StatusBadRequest); return }
    out, err := h.Svc.RenderTemplate(req.Template, req.Customer)
    if err!=nil{ w.WriteHeader(http.StatusInternalServerError); return }
    json.NewEncoder(w).Encode(map[string]string{"preview": out})
}
