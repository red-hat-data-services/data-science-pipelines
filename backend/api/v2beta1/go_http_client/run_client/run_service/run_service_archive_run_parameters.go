// Code generated by go-swagger; DO NOT EDIT.

package run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewRunServiceArchiveRunParams creates a new RunServiceArchiveRunParams object
// with the default values initialized.
func NewRunServiceArchiveRunParams() *RunServiceArchiveRunParams {
	var ()
	return &RunServiceArchiveRunParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewRunServiceArchiveRunParamsWithTimeout creates a new RunServiceArchiveRunParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewRunServiceArchiveRunParamsWithTimeout(timeout time.Duration) *RunServiceArchiveRunParams {
	var ()
	return &RunServiceArchiveRunParams{

		timeout: timeout,
	}
}

// NewRunServiceArchiveRunParamsWithContext creates a new RunServiceArchiveRunParams object
// with the default values initialized, and the ability to set a context for a request
func NewRunServiceArchiveRunParamsWithContext(ctx context.Context) *RunServiceArchiveRunParams {
	var ()
	return &RunServiceArchiveRunParams{

		Context: ctx,
	}
}

// NewRunServiceArchiveRunParamsWithHTTPClient creates a new RunServiceArchiveRunParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewRunServiceArchiveRunParamsWithHTTPClient(client *http.Client) *RunServiceArchiveRunParams {
	var ()
	return &RunServiceArchiveRunParams{
		HTTPClient: client,
	}
}

/*RunServiceArchiveRunParams contains all the parameters to send to the API endpoint
for the run service archive run operation typically these are written to a http.Request
*/
type RunServiceArchiveRunParams struct {

	/*RunID
	  The ID of the run to be archived.

	*/
	RunID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the run service archive run params
func (o *RunServiceArchiveRunParams) WithTimeout(timeout time.Duration) *RunServiceArchiveRunParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the run service archive run params
func (o *RunServiceArchiveRunParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the run service archive run params
func (o *RunServiceArchiveRunParams) WithContext(ctx context.Context) *RunServiceArchiveRunParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the run service archive run params
func (o *RunServiceArchiveRunParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the run service archive run params
func (o *RunServiceArchiveRunParams) WithHTTPClient(client *http.Client) *RunServiceArchiveRunParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the run service archive run params
func (o *RunServiceArchiveRunParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithRunID adds the runID to the run service archive run params
func (o *RunServiceArchiveRunParams) WithRunID(runID string) *RunServiceArchiveRunParams {
	o.SetRunID(runID)
	return o
}

// SetRunID adds the runId to the run service archive run params
func (o *RunServiceArchiveRunParams) SetRunID(runID string) {
	o.RunID = runID
}

// WriteToRequest writes these params to a swagger request
func (o *RunServiceArchiveRunParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param run_id
	if err := r.SetPathParam("run_id", o.RunID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
