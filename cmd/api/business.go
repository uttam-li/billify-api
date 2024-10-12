package main

import (
	"billify-api/internal/store"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateBusinessPayload struct {
	Name         string `json:"name" validate:"required,min=3,max=100"`
	GSTNo        string `json:"gstno" validate:"required,len=15,alphanum"`
	CompanyEmail string `json:"company_email" validate:"required,email"`
	CompanyPhone string `json:"company_phone" validate:"required,e164"`
	Address      string `json:"address" validate:"required,min=10,max=200"`
	City         string `json:"city" validate:"required,min=2,max=50"`
	ZipCode      string `json:"zip_code" validate:"required,len=6,numeric"`
	State        string `json:"state" validate:"required,min=2,max=50"`
	Country      string `json:"country" validate:"required,min=2,max=50"`
	BankName     string `json:"bank_name" validate:"required,min=3,max=100"`
	AccountNo    string `json:"account_no" validate:"required,numeric,min=9,max=18"`
	IFSC         string `json:"ifsc" validate:"required,len=11,alphanum"`
	BankBranch   string `json:"bank_branch" validate:"required,min=3,max=100"`
}

func (app *application) createBusinessHandler(w http.ResponseWriter, r *http.Request) {

	var payload CreateBusinessPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Get the user ID from the context
	user := r.Context().Value(userCtx).(*store.User)

	// Create a new business instance
	business := &store.Business{
		UserID:       user.ID,
		Name:         payload.Name,
		GSTNo:        payload.GSTNo,
		CompanyEmail: payload.CompanyEmail,
		CompanyPhone: payload.CompanyPhone,
		Address:      payload.Address,
		City:         payload.City,
		ZipCode:      payload.ZipCode,
		State:        payload.State,
		Country:      payload.Country,
		BankName:     payload.BankName,
		AccountNo:    payload.AccountNo,
		IFSC:         payload.IFSC,
		BankBranch:   payload.BankBranch,
	}

	if err := app.store.Business.Create(r.Context(), business); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Return the created business as a JSON response
	app.jsonResponse(w, http.StatusCreated, business)
}

func (app *application) getBusinessByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "busID")

	bussID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	business, err := app.store.Business.GetByID(r.Context(), bussID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, business); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getBusinessesByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(*store.User)

	businessess, err := app.store.Business.GetByUserID(r.Context(), user.ID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if len(businessess) == 0 {
		app.jsonResponse(w, http.StatusOK, "No businesses found for this user")
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, businessess); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type UpdateBusinessPayload struct {
	ID           uuid.UUID `json:"id" validate:"required,uuid"`
	Name         string    `json:"name" validate:"required,min=3,max=100"`
	GSTNo        string    `json:"gstno" validate:"required,len=15,alphanum"`
	CompanyEmail string    `json:"company_email" validate:"required,email"`
	CompanyPhone string    `json:"company_phone" validate:"required,e164"`
	Address      string    `json:"address" validate:"required,min=10,max=200"`
	City         string    `json:"city" validate:"required,min=2,max=50"`
	ZipCode      string    `json:"zip_code" validate:"required,len=6,numeric"`
	State        string    `json:"state" validate:"required,min=2,max=50"`
	Country      string    `json:"country" validate:"required,min=2,max=50"`
	BankName     string    `json:"bank_name" validate:"required,min=3,max=100"`
	AccountNo    string    `json:"account_no" validate:"required,numeric,min=9,max=18"`
	IFSC         string    `json:"ifsc" validate:"required,len=11,alphanum"`
	BankBranch   string    `json:"bank_branch" validate:"required,min=3,max=100"`
}

func (app *application) updateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdateBusinessPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Get the user ID from the context
	user := r.Context().Value(userCtx).(*store.User)

	// Create a new business instance
	business := &store.Business{
		ID:           payload.ID,
		UserID:       user.ID,
		Name:         payload.Name,
		GSTNo:        payload.GSTNo,
		CompanyEmail: payload.CompanyEmail,
		CompanyPhone: payload.CompanyPhone,
		Address:      payload.Address,
		City:         payload.City,
		ZipCode:      payload.ZipCode,
		State:        payload.State,
		Country:      payload.Country,
		BankName:     payload.BankName,
		AccountNo:    payload.AccountNo,
		IFSC:         payload.IFSC,
		BankBranch:   payload.BankBranch,
	}

	if err := app.store.Business.Update(r.Context(), business); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Return the created business as a JSON response
	app.jsonResponse(w, http.StatusCreated, business)
}

type BusinessDashboardRequest struct {
	BusId uuid.UUID `json:"busId" validate:"required,uuid"`
	From  time.Time `json:"from" validate:"required"`
	To    time.Time `json:"to" validate:"required"`
}

func (app *application) getBusinessesDashboardHandler(w http.ResponseWriter, r *http.Request) {

	var payload BusinessDashboardRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	data, err := app.store.Business.GetDashboard(r.Context(), payload.BusId, payload.From, payload.To)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
