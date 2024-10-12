package main

import (
	"billify-api/internal/store"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateCustomerPayload struct {
	BusinessID uuid.UUID `json:"bus_id" validate:"required,uuid"`
	GSTNo      string    `json:"gstno" validate:"required,len=15,alphanum"`
	Name       string    `json:"name" validate:"required,min=3,max=100"`
	Email      string    `json:"email" validate:"required,email"`
	Phone      string    `json:"phone" validate:"required,e164"`
	BAddress   string    `json:"b_address" validate:"required,min=10,max=200"`
	SAddress   string    `json:"s_address" validate:"required,min=10,max=200"`
}

func (app *application) createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCustomerPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create a new customer instance
	customer := &store.Customer{
		BusID:    payload.BusinessID,
		GSTNo:    payload.GSTNo,
		Name:     payload.Name,
		Email:    payload.Email,
		Phone:    payload.Phone,
		BAddress: payload.BAddress,
		SAddress: payload.SAddress,
	}

	if err := app.store.Customers.Create(r.Context(), customer); err != nil {
		log.Println(err)
		switch err {
		case store.ErrDuplicateCustomer:
			app.conflictResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.jsonResponse(w, http.StatusCreated, customer)
}


func (app *application) getCustomersByBusinessIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "busID")
	
	BusID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	customers, err := app.store.Customers.GetByBusID(r.Context(), BusID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	
	app.jsonResponse(w, http.StatusOK, customers)
}

type UpdateCustomerPayload struct {
    CustomerID uuid.UUID `json:"cust_id" validate:"required,uuid"`
    BusinessID uuid.UUID `json:"buss_id" validate:"required,uuid"`
    Name       string    `json:"name" validate:"required"`
    GSTNo      string    `json:"gstno" validate:"required"`
    Email      string    `json:"email" validate:"required,email"`
    Phone      string    `json:"phone" validate:"required"`
    BAddress   string    `json:"b_address" validate:"required"`
    SAddress   string    `json:"s_address" validate:"required"`
}

func (app *application) updateCustomerHandler(w http.ResponseWriter,r *http.Request) {
	var payload store.Customer
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Customers.Update(r.Context(), &payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    customerID, err := uuid.Parse(id)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    err = app.store.Customers.Delete(r.Context(), customerID)
    if err != nil {
        switch err {
        case store.ErrNotFound:
            app.notFoundResponse(w, r, err)
        default:
            app.internalServerError(w, r, err)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}