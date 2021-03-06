// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// ShareElement share element
// swagger:model ShareElement
type ShareElement struct {

	// content filter
	ContentFilter map[string]interface{} `json:"content_filter,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// element watch
	ElementWatch bool `json:"element_watch,omitempty"`

	// entries
	Entries []*ShareEntry `json:"entries"`

	// links
	Links interface{} `json:"links,omitempty"`

	// repository url
	RepositoryURL string `json:"repository_url,omitempty"`

	// repositoryid
	Repositoryid string `json:"repositoryid,omitempty"`

	// share owner
	ShareOwner string `json:"share_owner,omitempty"`

	// share scope
	ShareScope string `json:"share_scope,omitempty"`

	// users number
	UsersNumber int64 `json:"users_number,omitempty"`
}

// Validate validates this share element
func (m *ShareElement) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEntries(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ShareElement) validateEntries(formats strfmt.Registry) error {

	if swag.IsZero(m.Entries) { // not required
		return nil
	}

	for i := 0; i < len(m.Entries); i++ {
		if swag.IsZero(m.Entries[i]) { // not required
			continue
		}

		if m.Entries[i] != nil {
			if err := m.Entries[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("entries" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ShareElement) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ShareElement) UnmarshalBinary(b []byte) error {
	var res ShareElement
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
