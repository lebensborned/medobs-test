package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lebensborned/medobs-test/server/api"
	"github.com/lebensborned/medobs-test/store/tokeninfo"
	"github.com/lebensborned/medobs-test/utils"
	"golang.org/x/crypto/bcrypt"
)

// GetTokens generates a pair of access && refresh token
func (srv *Server) GetTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := vars["guid"]; !ok {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Parameter not specified",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	td, err := utils.CreateTokenPair(vars["guid"])
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	hashedRt, err := bcrypt.GenerateFromPassword([]byte(td.RefreshToken), bcrypt.DefaultCost)
	insModel := tokeninfo.Model{
		GUID:         vars["guid"],
		RefreshToken: string(hashedRt),
	}
	err = insModel.Save(srv.store)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(td.RefreshToken))
	tokens := map[string]string{
		"access_token":  td.AccessToken,
		"refresh_token": sEnc,
	}
	response := api.NewResponse(tokens)
	response.WriteResponse(w)

	return
}

// RefreshTokens performs refresh operation to pair of access && refresh tokens
func (srv *Server) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  err.Error(),
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	tokens := map[string]string{}
	err = json.Unmarshal(body, &tokens)
	accessToken, ok := tokens["access_token"]
	if !ok {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Parameter not specified",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	refreshToken, ok := tokens["refresh_token"]
	if !ok {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Parameter not specified",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	sDec, err := base64.StdEncoding.DecodeString(refreshToken)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Refresh token was modified",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}

	// check refresh token
	token, err := jwt.Parse(string(sDec), func(token *jwt.Token) (interface{}, error) {
		// check token method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	// check is refresh token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  err.Error(),
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	// if error => the refresh token have expired
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInvalidToken,
			Msg:  "Refresh token have expired",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	_, ok = claims["refresh_uuid"].(string)
	if !ok {
		newErr := api.Error{
			Code: api.CodeInvalidToken,
			Msg:  "Invalid token",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		newErr := api.Error{
			Code: api.CodeInvalidToken,
			Msg:  "Invalid token",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	result, err := tokeninfo.FindByGUID(srv.store, userID)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.RefreshToken), sDec)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Attempt to reuse tokens",
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	// check access and refresh token compatibility
	acToken, _ := jwt.Parse(accessToken, func(acToken *jwt.Token) (interface{}, error) {
		if _, ok := acToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", acToken.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	// check is access token valid?
	if _, ok := acToken.Claims.(jwt.Claims); !ok && !acToken.Valid {
		newErr := api.Error{
			Code: api.CodeInvalidToken,
			Msg:  "Invalid token",
		}
		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	atClaims, ok := acToken.Claims.(jwt.MapClaims)
	if ok && acToken.Valid {
		_, ok := atClaims["access_uuid"].(string)
		if !ok {
			newErr := api.Error{
				Code: api.CodeInvalidToken,
				Msg:  "Invalid token",
			}
			response := newErr.GetResponse()
			response.WriteResponse(w)
			return
		}
	}
	if claims["refresh_uuid"].(string) != atClaims["access_uuid"].(string) {
		newErr := api.Error{
			Code: api.CodeInvalidRequest,
			Msg:  "Invalid token pair",
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}

	// create new pairs of refresh and access tokens
	ts, createErr := utils.CreateTokenPair(userID)
	if createErr != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	hashedRt, err := bcrypt.GenerateFromPassword([]byte(ts.RefreshToken), bcrypt.DefaultCost)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	result.RefreshToken = string(hashedRt)
	err = result.Save(srv.store)
	if err != nil {
		newErr := api.Error{
			Code: api.CodeInternalError,
			Msg:  err.Error(),
		}

		response := newErr.GetResponse()
		response.WriteResponse(w)
		return
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(ts.RefreshToken))
	tokens = map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": sEnc,
	}
	response := api.NewResponse(tokens)
	response.WriteResponse(w)
	return
}
