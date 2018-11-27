package payment

import (
	"Rave-go/rave"
	"Rave-go/rave/helper"
	"go/types"
	
)


type CardCharge interface {
	ChargeCard(data CardChargeData) (error error, response map[string]interface{})
}

type CardValidate interface {
	ValidateCard(data CardValidateData) (error error, response map[string]interface{})
}

type CardVerify interface {
	VerifyCard(data CardVerifyData) (error error, response map[string]interface{})
}

type CardTokenized interface {
	TokenizedCharge(data SaveCardChargeData) (error error, response map[string]interface{})
}

type PreauthCharge interface {
	ChargePreauth(data CardChargeData) (error error, response map[string]interface{})
}

type PreauthCapture interface {
	CapturePreauth(data CardValidateData) (error error, response map[string]interface{})
}

type PreauthRefundOrVoid interface {
	RefundOrVoidPreauth(data CardVerifyData) (error error, response map[string]interface{})
}

type CardInterface interface {
	CardCharge
	CardValidate
	CardVerify
	CardTokenized
	PreauthCharge
	PreauthCapture
	PreauthRefundOrVoid
}

type CardChargeData struct {
	Cardno               string         `json:"cardno"`
	Cvv                  string         `json:"cvv"`
	Expirymonth          string         `json:"expirymonth"`
	Expiryyear           string         `json:"expiryyear"`
	Pin                  string         `json:"pin"`
	Amount               float64        `json:"amount"`
	Currency             string         `json:"currency"`
	Country              string         `json:"country"`
	CustomerPhone        string         `json:"customer_phone"`
	Firstname            string         `json:"firstname"`
	Lastname             string         `json:"lastname"`
	Email                string         `json:"email"`
	Ip                   string         `json:"IP"`
	Txref		         string	        `json:"txRef"`
	RedirectUrl          string         `json:"redirect_url"`
	Subaccounts          types.Slice    `json:"subaccounts"`
	DeviceFingerprint    string         `json:"device_fingerprint"`
	Meta                 types.Slice    `json:"meta"`
	SuggestedAuth        string         `json:"suggested_auth"`
	BillingZip           string         `json:"billingzip"`
	BillingCity          string         `json:"billingcity"`
	BillingAddress       string         `json:"billingaddress"`
	BillingState         string         `json:"billingstate"`
	BillingCountry       string         `json:"billingcountry"`
	Chargetype		     string	        `json:"charge_type"`

}

type CardValidateData struct {
	Reference	   string	      `json:"transaction_reference"`
	Otp		       string	      `json:"otp"`
	PublicKey      string         `json:"PBFPubKey"`
}

type CardVerifyData struct {
	Reference	   string	      `json:"txref"`
	Amount	       float64	      `json:"amount"`
	Currency       string         `json:"currency"`
	SecretKey      string         `json:"SECKEY"`
}

type SaveCardChargeData struct {
	SecretKey      string         `json:"SECKEY"`
	Currency       string         `json:"currency"`
	Token          string         `json:"token"`
	Country        string         `json:"country"`
	Amount	       float64	      `json:"amount"`
	Email          string         `json:"email"`
	Firstname      string         `json:"firstname"`
	Lastname       string         `json:"lastname"`
	Ip             string         `json:"IP"`
	Txref		   string	      `json:"txRef"`
}

type PreauthCaptureData struct {
	SecretKey      string         `json:"SECKEY"`
	Amount	       float64	      `json:"amount"`
	Flwref	       string	      `json:"flwRef"`

}

type PreauthRefundData struct {
	Flwref	       string	          `json:"ref"`
	Action	       string	          `json:"action"`
	SecretKey      string             `json:"SECKEY"`
}

type Card struct {
	rave.Rave
}

func (c Card) ChargeCard(data CardChargeData) (error error, response map[string]interface{}) {
	var url string
	// if (data.Txref == "") {
	// 	data.Txref = GenerateRef()
	// }
	chargeJSON := helper.MapToJSON(data)
	encryptedChargeData := c.Encrypt(string(chargeJSON[:]))
	queryParam := map[string]interface{}{
        "PBFPubKey": c.GetPublicKey(),
        "client": encryptedChargeData,
        "alg": "3DES-24",
	}
	// postData := SetupCharge(data)
	
	if data.Chargetype == "preauth" {
		url = c.GetBaseURL() + c.GetEndpoint("preauth", "charge")
	} else {
		url = c.GetBaseURL() + c.GetEndpoint("card", "charge")
	}

	err, response := helper.MakePostRequest(queryParam, url)
	if err != nil {
		return err, noresponse
	}
	suggestedAuth := response["data"].(map[string]interface{})["suggested_auth"]
	if (suggestedAuth == "PIN") {
		data.SuggestedAuth = "PIN"
		chargeJSON = helper.MapToJSON(data)
		encryptedChargeData = c.Encrypt(string(chargeJSON[:]))
		queryParam = map[string]interface{}{
			"PBFPubKey": c.GetPublicKey(),
			"client": encryptedChargeData,
			"alg": "3DES-24",
		}
		err, response = helper.MakePostRequest(queryParam, url)
		if err != nil {
			return err, noresponse
		}
	} else if (suggestedAuth == "AVS_VBVSECURECODE") {
		data.SuggestedAuth = "AVS_VBVSECURECODE"
		chargeJSON = helper.MapToJSON(data)
		encryptedChargeData = c.Encrypt(string(chargeJSON[:]))
		queryParam = map[string]interface{}{
			"PBFPubKey": c.GetPublicKey(),
			"client": encryptedChargeData,
			"alg": "3DES-24",
		}
		err, response = helper.MakePostRequest(queryParam, url)
		if err != nil {
			return err, noresponse
		}

	}

	return nil, response

}

func (c Card) ValidateCard(data CardValidateData) (error error, response map[string]interface{}) {
	data.PublicKey = c.GetPublicKey()
	url := c.GetBaseURL() + c.GetEndpoint("card", "validate")
	err, response := helper.MakePostRequest(data, url)
	if err != nil {
		return err, noresponse
	}
	return nil, response

}

func (c Card) VerifyCard(data CardVerifyData) (error error, response map[string]interface{}) {
	data.SecretKey = c.GetSecretKey()
	url := c.GetBaseURL() + c.GetEndpoint("card", "verify")
	err, response := helper.MakePostRequest(data, url)
	
	transactionRef := response["data"].(map[string]interface{})["txref"].(string)
	status := response["status"].(string) 
	chargeCode := response["data"].(map[string]interface{})["chargecode"].(string)
	amount := response["data"].(map[string]interface{})["chargedamount"].(float64)
	currency := response["data"].(map[string]interface{})["currency"].(string)
	
	transactionReference := data.Reference
	currencyCode := data.Currency
	chargedAmount := data.Amount

	err = VerifyTransactionReference(transactionRef, transactionReference)
	err = VerifySuccessMessage(status)
	err = VerifyChargeResponse(chargeCode)
	err = VerifyCurrencyCode(currency, currencyCode)
	err = VerifyChargedAmount(amount, chargedAmount)
	
	if err != nil {
		return err, noresponse
	}
	return nil, response

}

func (c Card) TokenizedCharge(data SaveCardChargeData) (error error, response map[string]interface{}) {
	data.SecretKey = c.GetSecretKey()
	url := c.GetBaseURL() + c.GetEndpoint("card", "chargeSavedCard")
	err, response := helper.MakePostRequest(data, url)
	if err != nil {
		return err, noresponse
	}
	
	return nil, response

}

func (c Card) ChargePreauth(data CardChargeData) (error error, response map[string]interface{}) {
	data.Chargetype = "preauth"
	err, response := c.ChargeCard(data)
	if err != nil {
		return err, noresponse
	}

	return nil, response

}

func (c Card) CapturePreauth(data PreauthCaptureData) (error error, response map[string]interface{}) {
	data.SecretKey = c.GetSecretKey()
	url := c.GetBaseURL() + c.GetEndpoint("preauth", "capture")
	err, response := helper.MakePostRequest(data, url)
	if err != nil {
		return err, noresponse
	}
	return nil, response

}

func (c Card) RefundOrVoidPreauth(data PreauthRefundData) (error error, response map[string]interface{}) {
	data.SecretKey = c.GetSecretKey()
	url := c.GetBaseURL() + c.GetEndpoint("preauth", "refundorvoid")
	err, response := helper.MakePostRequest(data, url)
	if err != nil {
		return err, noresponse
	}
	
	return nil, response

}