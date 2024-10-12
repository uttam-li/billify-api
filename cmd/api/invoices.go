package main

import (
	"billify-api/internal/store"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InvoiceItemPayload struct {
	ID        uuid.UUID `json:"id"`
	ProdID    uuid.UUID `json:"prod_id" validate:"required,uuid"`
	TaxRate   float64   `json:"tax_rate"`
	Quantity  int       `json:"quantity" validate:"required"`
	UnitPrice float64   `json:"unit_price" validate:"required"`
}

type InvoicePayload struct {
	ID          uuid.UUID                  `json:"id" `
	InvNo       int64                      `json:"inv_no" validate:"required"`
	BusID       uuid.UUID                  `json:"bus_id" validate:"required,uuid"`
	CustID      uuid.UUID                  `json:"cust_id" validate:"required,uuid"`
	TotalAmount float64                    `json:"total_amount" validate:"required"`
	InvDate     time.Time                  `json:"inv_date" validate:"required"`
	DueDate     time.Time                  `json:"due_date" validate:"required"`
	IsPaid      bool                       `json:"is_paid"`
	PaidDate    *time.Time                 `json:"paid_date,omitempty"`
	Items       []InvoiceItemPayload `json:"items" validate:"required,dive"`
}

func (app *application) createInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	var payload InvoicePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	invoice := &store.Invoice{
		InvNo:       payload.InvNo,
		BusID:       payload.BusID,
		CustID:      payload.CustID,
		TotalAmount: payload.TotalAmount,
		InvDate:     payload.InvDate,
		DueDate:     payload.DueDate,
		IsPaid:      payload.IsPaid,
		PaidDate:    payload.PaidDate,
	}

	err := app.store.Invoices.Create(r.Context(), invoice)
	if err != nil {
		if err == store.ErrDuplicateInvoice {
			app.conflictResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	// var items []*store.InvoiceItem
	for _, itemPayload := range payload.Items {
		item := store.InvoiceItem{
			InvID:     invoice.ID,
			ProdID:    itemPayload.ProdID,
			Quantity:  itemPayload.Quantity,
			UnitPrice: itemPayload.UnitPrice,
		}
		err = app.store.InvoiceItems.Create(r.Context(), &item)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		// items = append(items, &item)
	}

	w.WriteHeader(http.StatusCreated)
}

func (app *application) getNextInvoiceNumberHandler(w http.ResponseWriter, r *http.Request) {
	busID := chi.URLParam(r, "busID")

	businessID, err := uuid.Parse(busID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	nextInvNo, err := app.store.Invoices.GetNextInvoiceNumber(r.Context(), businessID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	response := map[string]int64{"next_invoice_number": nextInvNo}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *application) updateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	var payload InvoicePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	invoice := &store.Invoice{
		ID:          payload.ID,
		InvNo:       payload.InvNo,
		BusID:       payload.BusID,
		CustID:      payload.CustID,
		TotalAmount: payload.TotalAmount,
		InvDate:     payload.InvDate,
		DueDate:     payload.DueDate,
		IsPaid:      payload.IsPaid,
		PaidDate:    payload.PaidDate,
	}

	err := app.store.Invoices.Update(r.Context(), invoice)
	if err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	var items []*store.InvoiceItem
	for _, itemPayload := range payload.Items {
		item := &store.InvoiceItem{
			ID:        itemPayload.ID,
			InvID:     invoice.ID,
			ProdID:    itemPayload.ProdID,
			Quantity:  itemPayload.Quantity,
			UnitPrice: itemPayload.UnitPrice,
		}
		items = append(items, item)
	}

	err = app.store.InvoiceItems.UpdateAll(r.Context(), invoice.ID, items)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.store.Invoices.Delete(r.Context(), invoiceID)
	if err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getInvoiceByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	invoice, err := app.store.Invoices.GetByID(r.Context(), invoiceID)
	if err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	items, err := app.store.InvoiceItems.GetByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	response := store.InvoiceResponse{
		ID:          invoice.ID,
		InvNo:       invoice.InvNo,
		BusID:       invoice.BusID,
		CustID:      invoice.CustID,
		TotalAmount: invoice.TotalAmount,
		InvDate:     invoice.InvDate,
		DueDate:     invoice.DueDate,
		IsPaid:      invoice.IsPaid,
		PaidDate:    invoice.PaidDate,
		CreatedAt:   invoice.CreatedAt,
		Items:       items,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getInvoicesByBusinessIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "busID")

	busID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	invoices, err := app.store.Invoices.GetByBusID(r.Context(), busID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var response []store.InvoiceResponse
	for _, invoice := range invoices {
		items, err := app.store.InvoiceItems.GetByInvoiceID(r.Context(), invoice.ID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		response = append(response, store.InvoiceResponse{
			ID:          invoice.ID,
			InvNo:       invoice.InvNo,
			BusID:       invoice.BusID,
			CustID:      invoice.CustID,
			TotalAmount: invoice.TotalAmount,
			InvDate:     invoice.InvDate,
			DueDate:     invoice.DueDate,
			IsPaid:      invoice.IsPaid,
			PaidDate:    invoice.PaidDate,
			CreatedAt:   invoice.CreatedAt,
			Items:       items,
		})
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) updateInvoiceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Invoices.UpdateStatus(r.Context(), invoiceID); err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}
}

func (app *application) getInvoiceAsPDFHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "invID")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	invoice, err := app.store.Invoices.GetByID(r.Context(), invoiceID)
	if err != nil {
		if err == store.ErrNotFound {
			app.notFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	items, err := app.store.InvoiceItems.GetByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	customer, err := app.store.Customers.GetByID(r.Context(), invoice.CustID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	business, err := app.store.Business.GetByID(r.Context(), invoice.BusID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	product, err := app.store.Products.GetByBusID(r.Context(), invoice.BusID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	pdfData, err := app.pdf.GenerateInvoicePDF(business, invoice, customer, items, product)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}
