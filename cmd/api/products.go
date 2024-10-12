package main

import (
	"billify-api/internal/store"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateProductPayload struct {
	BusID   uuid.UUID `json:"bus_id" validate:"required,uuid"`
	Name    string    `json:"name" validate:"required,min=3,max=100"`
	Price   float64   `json:"price" validate:"required"`
	TaxRate float64   `json:"tax_rate" validate:"required"`
	Unit    string    `json:"unit" validate:"required"`
	HSNCode string    `json:"hsn_code" validate:"required"`
}

func (app *application) createProductHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateProductPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &store.Product{
		BusID:   payload.BusID,
		Name:    payload.Name,
		Price:   payload.Price,
		TaxRate: payload.TaxRate,
		Unit:    payload.Unit,
		HSNCode: payload.HSNCode,
	}

	if err := app.store.Products.Create(r.Context(), product); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type UpdateProductPayload struct {
	ID      uuid.UUID `json:"id" validate:"required,uuid"`
	Name    string    `json:"name" validate:"required,min=3,max=100"`
	Price   float64   `json:"price" validate:"required,numeric"`
	TaxRate float64   `json:"tax_rate" validate:"required,numeric"`
	Unit    string    `json:"unit" validate:"required"`
	HSNCode string    `json:"hsn_code" validate:"required"`
}

func (app *application) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdateProductPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &store.Product{
		ID:      payload.ID,
		Name:    payload.Name,
		Price:   payload.Price,
		TaxRate: payload.TaxRate,
		Unit:    payload.Unit,
		HSNCode: payload.HSNCode,
	}

	if err := app.store.Products.Update(r.Context(), product); err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	productID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Products.Delete(r.Context(), productID); err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getProductsByBusinessIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "busID")

	busID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	products, err := app.store.Products.GetByBusID(r.Context(), busID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, products)
}

