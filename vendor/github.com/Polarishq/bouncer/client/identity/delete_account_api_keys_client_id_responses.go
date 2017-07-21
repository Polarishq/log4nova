package identity

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/Polarishq/bouncer/models"
)

// DeleteAccountAPIKeysClientIDReader is a Reader for the DeleteAccountAPIKeysClientID structure.
type DeleteAccountAPIKeysClientIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteAccountAPIKeysClientIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewDeleteAccountAPIKeysClientIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewDeleteAccountAPIKeysClientIDDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteAccountAPIKeysClientIDOK creates a DeleteAccountAPIKeysClientIDOK with default headers values
func NewDeleteAccountAPIKeysClientIDOK() *DeleteAccountAPIKeysClientIDOK {
	return &DeleteAccountAPIKeysClientIDOK{}
}

/*DeleteAccountAPIKeysClientIDOK handles this case with default header values.

API Keys Deleted
*/
type DeleteAccountAPIKeysClientIDOK struct {
}

func (o *DeleteAccountAPIKeysClientIDOK) Error() string {
	return fmt.Sprintf("[DELETE /account/apiKeys/{client_id}][%d] deleteAccountApiKeysClientIdOK ", 200)
}

func (o *DeleteAccountAPIKeysClientIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteAccountAPIKeysClientIDDefault creates a DeleteAccountAPIKeysClientIDDefault with default headers values
func NewDeleteAccountAPIKeysClientIDDefault(code int) *DeleteAccountAPIKeysClientIDDefault {
	return &DeleteAccountAPIKeysClientIDDefault{
		_statusCode: code,
	}
}

/*DeleteAccountAPIKeysClientIDDefault handles this case with default header values.

Unexpected error
*/
type DeleteAccountAPIKeysClientIDDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the delete account API keys client ID default response
func (o *DeleteAccountAPIKeysClientIDDefault) Code() int {
	return o._statusCode
}

func (o *DeleteAccountAPIKeysClientIDDefault) Error() string {
	return fmt.Sprintf("[DELETE /account/apiKeys/{client_id}][%d] DeleteAccountAPIKeysClientID default  %+v", o._statusCode, o.Payload)
}

func (o *DeleteAccountAPIKeysClientIDDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}