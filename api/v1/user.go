package apiv1

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/dankobgd/ecommerce-shop/model"
	"github.com/dankobgd/ecommerce-shop/utils/locale"
	"github.com/dankobgd/ecommerce-shop/utils/pagination"
	"github.com/go-chi/chi"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	msgInvalidToken         = &i18n.Message{ID: "model.access_token_verify.json.app_error", Other: "token is invalid or has already expired"}
	msgUserFromJSON         = &i18n.Message{ID: "api.user.create_user.json.app_error", Other: "could not decode user json data"}
	msgRefreshTokenFromJSON = &i18n.Message{ID: "api.user.create_user.json.app_error", Other: "could not decode token json data"}
	msgInvalidEmail         = &i18n.Message{ID: "api.user.sendUserVerificationEmail.email.app_error", Other: "invalid email provided"}
	msgInvalidPassword      = &i18n.Message{ID: "api.user.reset_user_password.password.app_error", Other: "invalid password provided"}
	msgUserURLParams        = &i18n.Message{ID: "api.user.delete_user.app_error", Other: "invalid user_id url param"}
	msgUserMultiPartErr     = &i18n.Message{ID: "api.user.create_user.app_error", Other: "invalid user form data"}
	msgAddressFromJSON      = &i18n.Message{ID: "api.user.delete_user.app_error", Other: "could not parse address json data"}
	msgAddressPatchFromJSON = &i18n.Message{ID: "api.user.delete_user.app_error", Other: "could not parse address patch data"}
	msgDeleteUserAddress    = &i18n.Message{ID: "api.user.delete_user.app_error", Other: "could not delete address"}
	msgUserAvatarMultipart  = &i18n.Message{ID: "api.user.upload_user_avatar.app_error", Other: "could not parse avatar multipart file"}
	msgUpdateProfile        = &i18n.Message{ID: "api.user.update_profile.app_error", Other: "could not update user profile"}
	msgGetUserOrders        = &i18n.Message{ID: "api.user.get_user_orders.app_error", Other: "could not get user orders"}
	msgWishlistParamErr     = &i18n.Message{ID: "api.user.wishlist.app_error", Other: "invalid wishlist product_id"}
)

// InitUser inits the user routes
func InitUser(a *API) {
	a.Routes.Users.Get("/", a.AdminSessionRequired(a.getUsers))
	a.Routes.Users.Get("/me", a.SessionRequired(a.currentUser))
	a.Routes.Users.Post("/new", a.createUser)
	a.Routes.Users.Post("/", a.signup)
	a.Routes.Users.Post("/login", a.login)
	a.Routes.Users.Post("/logout", a.SessionRequired(a.logout))
	a.Routes.Users.Delete("/bulk", a.AdminSessionRequired(a.deleteUsers))
	a.Routes.Users.Post("/token/refresh", a.refresh)
	a.Routes.Users.Post("/email/verify", a.verifyUserEmail)
	a.Routes.Users.Post("/email/verify/send", a.sendVerificationEmail)
	a.Routes.Users.Post("/password/reset", a.resetUserPassword)
	a.Routes.Users.Post("/password/reset/send", a.sendPasswordResetEmail)
	a.Routes.Users.Patch("/profile", a.SessionRequired(a.updateProfile))
	a.Routes.Users.Put("/password", a.SessionRequired(a.changeUserPassword))
	a.Routes.Users.Post("/avatar", a.SessionRequired(a.uploadUserAvatar))
	a.Routes.Users.Patch("/avatar", a.SessionRequired(a.deleteUserAvatar))
	a.Routes.Users.Post("/addresses", a.SessionRequired(a.createUserAddress))
	a.Routes.Users.Get("/addresses", a.SessionRequired(a.getUserAddresses))
	a.Routes.Users.Get("/addresses/{address_id:[A-Za-z0-9]+}", a.SessionRequired(a.getUserAddress))
	a.Routes.Users.Patch("/addresses/{address_id:[A-Za-z0-9]+}", a.SessionRequired(a.updateUserAddress))
	a.Routes.Users.Delete("/addresses/{address_id:[A-Za-z0-9]+}", a.SessionRequired(a.deleteUserAddress))
	a.Routes.Users.Post("/wishlist", a.SessionRequired(a.createWishlist))
	a.Routes.Users.Get("/wishlist", a.SessionRequired(a.getWishlist))
	a.Routes.Users.Delete("/wishlist/{product_id:[A-Za-z0-9]+}", a.SessionRequired(a.deleteWishlist))
	a.Routes.Users.Delete("/wishlist/clear", a.SessionRequired(a.clearWishlist))

	a.Routes.User.Get("/", a.SessionRequired(a.getUser))
	a.Routes.User.Patch("/", a.AdminSessionRequired(a.update))
	a.Routes.User.Delete("/", a.SessionRequired(a.deleteUser))
	a.Routes.User.Get("/orders", a.SessionRequired(a.getUserOrders))
}

func (a *API) currentUser(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	user, err := a.app.GetUserByID(uid)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(model.FileUploadSizeLimit); err != nil {
		respondError(w, model.NewAppErr("createUser", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserMultiPartErr, http.StatusInternalServerError, nil))
		return
	}

	mpf := r.MultipartForm
	model.SchemaDecoder.IgnoreUnknownKeys(true)

	u := &model.User{}
	if err := model.SchemaDecoder.Decode(u, mpf.Value); err != nil {
		respondError(w, model.NewAppErr("createUser", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserMultiPartErr, http.StatusInternalServerError, nil))
		return
	}

	var fh *multipart.FileHeader
	if len(mpf.File["avatar_url"]) > 0 {
		fh = mpf.File["avatar_url"][0]
	}

	user, uErr := a.app.CreateUser(u, fh)
	if uErr != nil {
		respondError(w, uErr)
		return
	}
	respondJSON(w, http.StatusCreated, user)
}

func (a *API) signup(w http.ResponseWriter, r *http.Request) {
	u, e := model.UserFromJSON(r.Body)
	if e != nil {
		respondError(w, model.NewAppErr("signup", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserFromJSON, http.StatusInternalServerError, nil))
		return
	}

	user, err := a.app.Signup(u)
	if err != nil {
		respondError(w, err)
		return
	}

	tokenMeta, err := a.app.IssueTokens(user)
	if err != nil {
		respondError(w, err)
	}
	if err := a.app.SaveAuth(user.ID, tokenMeta); err != nil {
		respondError(w, err)
	}
	a.app.AttachSessionCookies(w, tokenMeta)
	respondJSON(w, http.StatusCreated, user)
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	u, e := model.UserLoginFromJSON(r.Body)
	if e != nil {
		respondError(w, model.NewAppErr("login", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserFromJSON, http.StatusInternalServerError, nil))
		return
	}

	user, err := a.app.Login(u)
	if err != nil {
		respondError(w, err)
		return
	}

	tokenMeta, err := a.app.IssueTokens(user)
	if err != nil {
		respondError(w, err)
	}
	if err := a.app.SaveAuth(user.ID, tokenMeta); err != nil {
		respondError(w, err)
	}
	a.app.AttachSessionCookies(w, tokenMeta)
	respondJSON(w, http.StatusOK, user)
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	a.app.DeleteSessionCookies(w)
	ad, err := a.app.ExtractTokenMetadata(r)
	if err != nil {
		respondError(w, err)
		return
	}
	deleted, err := a.app.DeleteAuth(ad.AccessUUID)
	if err != nil || deleted == 0 {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	rt, e := model.RefreshTokenFromJSON(r.Body)
	if e != nil {
		respondError(w, model.NewAppErr("refresh", model.ErrInternal, locale.GetUserLocalizer("en"), msgRefreshTokenFromJSON, http.StatusInternalServerError, nil))
		return
	}

	meta, err := a.app.RefreshToken(rt)
	if err != nil {
		respondError(w, err)
		return
	}

	a.app.AttachSessionCookies(w, meta)
	respondOK(w)
}

func (a *API) sendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	props := model.MapStrStrFromJSON(r.Body)
	email := props["email"]
	email = model.NormalizeEmail(email)

	if len(email) == 0 || !model.IsValidEmail(email) {
		respondError(w, model.NewAppErr("api.sendVerificationEmail", model.ErrInvalid, locale.GetUserLocalizer("en"), msgInvalidEmail, http.StatusBadRequest, nil))
		return
	}

	user, err := a.app.GetUserByEmail(email)
	if err != nil {
		// don't leak whether email is valid and exists - maybe for demonstration return some err
		respondOK(w)
		return
	}

	if err := a.app.SendVerificationEmail(user, email); err != nil {
		// don't leak whether email is valid and exists - maybe for demonstration return some err
		respondOK(w)
		return
	}

	respondOK(w)
}

func (a *API) verifyUserEmail(w http.ResponseWriter, r *http.Request) {
	props := model.MapStrStrFromJSON(r.Body)
	token := props["token"]

	if len(token) == 0 {
		respondError(w, model.NewAppErr("api.sendVerificationEmail", model.ErrInvalid, locale.GetUserLocalizer("en"), msgInvalidToken, http.StatusBadRequest, nil))
		return
	}

	if err := a.app.VerifyUserEmail(token); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) sendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {
	props := model.MapStrStrFromJSON(r.Body)
	email := props["email"]
	email = model.NormalizeEmail(email)

	if len(email) == 0 || !model.IsValidEmail(email) {
		respondError(w, model.NewAppErr("api.sendPasswordResetEmail", model.ErrInvalid, locale.GetUserLocalizer("en"), msgInvalidEmail, http.StatusBadRequest, nil))
		return
	}

	if err := a.app.SendPasswordResetEmail(email); err != nil {
		respondError(w, err)
		return
	}

	respondOK(w)
}

func (a *API) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	props := model.MapStrStrFromJSON(r.Body)
	token := props["token"]
	newPassword := props["password"]

	if err := a.app.ResetUserPassword(token, newPassword); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) update(w http.ResponseWriter, r *http.Request) {
	uid, e := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("update", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserMultiPartErr, http.StatusInternalServerError, nil))
		return
	}

	if err := r.ParseMultipartForm(model.FileUploadSizeLimit); err != nil {
		respondError(w, model.NewAppErr("update", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserAvatarMultipart, http.StatusInternalServerError, nil))
		return
	}

	mpf := r.MultipartForm
	model.SchemaDecoder.IgnoreUnknownKeys(true)

	patch := &model.UserPatch{}
	if err := model.SchemaDecoder.Decode(patch, mpf.Value); err != nil {
		respondError(w, model.NewAppErr("update", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserMultiPartErr, http.StatusInternalServerError, nil))
		return
	}

	var avatar *multipart.FileHeader
	if len(mpf.File["avatar_url"]) > 0 {
		avatar = mpf.File["avatar_url"][0]
	}

	uuser, pErr := a.app.PatchUser(uid, patch, avatar)
	if e != nil {
		respondError(w, pErr)
		return
	}

	respondJSON(w, http.StatusOK, uuser)
}

func (a *API) updateProfile(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	patch, err := model.UserPatchFromJSON(r.Body)
	if err != nil {
		respondError(w, model.NewAppErr("updateProfile", model.ErrInternal, locale.GetUserLocalizer("en"), msgUpdateProfile, http.StatusInternalServerError, nil))
		return
	}

	user, pErr := a.app.PatchUserProfile(uid, patch)
	if pErr != nil {
		respondError(w, pErr)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (a *API) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	props := model.MapStrStrFromJSON(r.Body)
	oldPassword := props["old_password"]
	newPassword := props["new_password"]
	confirmPassword := props["confirm_password"]

	if len(oldPassword) == 0 || len(newPassword) == 0 || len(confirmPassword) == 0 || newPassword != confirmPassword {
		respondError(w, model.NewAppErr("api.changeUserPassword", model.ErrInvalid, locale.GetUserLocalizer("en"), msgInvalidPassword, http.StatusBadRequest, nil))
		return
	}

	if err := a.app.ChangeUserPassword(uid, oldPassword, newPassword); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) getUser(w http.ResponseWriter, r *http.Request) {
	uid, e := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("getUser", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserURLParams, http.StatusInternalServerError, nil))
		return
	}

	user, err := a.app.GetUserByID(uid)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (a *API) getUsers(w http.ResponseWriter, r *http.Request) {
	pages := pagination.NewFromRequest(r)
	users, err := a.app.GetUsers(pages.Limit(), pages.Offset())
	if err != nil {
		respondError(w, err)
		return
	}

	totalCount := -1
	if len(users) > 0 {
		totalCount = users[0].TotalCount
	}
	pages.SetData(users, totalCount)

	respondJSON(w, http.StatusOK, pages)
}

func (a *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	uid, e := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("deleteUser", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserURLParams, http.StatusInternalServerError, nil))
		return
	}

	if err := a.app.DeleteUser(uid); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) deleteUsers(w http.ResponseWriter, r *http.Request) {
	ids := model.IntSliceFromJSON(r.Body)

	if err := a.app.DeleteUsers(ids); err != nil {
		respondError(w, err)
		return
	}

	respondOK(w)
}

func (a *API) uploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	if e := r.ParseMultipartForm(model.FileUploadSizeLimit); e != nil {
		respondError(w, model.NewAppErr("uploadUserAvatar", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserAvatarMultipart, http.StatusInternalServerError, nil))
		return
	}

	f, fh, err := r.FormFile("avatar")
	if err != nil {
		respondError(w, model.NewAppErr("uploadUserAvatar", model.ErrInternal, locale.GetUserLocalizer("en"), msgUserAvatarMultipart, http.StatusInternalServerError, nil))
		return
	}
	defer f.Close()

	url, publicID, uErr := a.app.UploadUserAvatar(uid, f, fh)
	if uErr != nil {
		respondError(w, uErr)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{"avatar_url": url, "avatar_public_id": publicID})
}

func (a *API) deleteUserAvatar(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	user, err := a.app.GetUserByID(uid)
	if err != nil {
		respondError(w, err)
	}

	if err := a.app.DeleteUserAvatar(uid, *user.AvatarPublicID); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) createUserAddress(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	addr, e := model.AddressFromJSON(r.Body)
	if e != nil {
		respondError(w, model.NewAppErr("createUserAddress", model.ErrInternal, locale.GetUserLocalizer("en"), msgAddressFromJSON, http.StatusInternalServerError, nil))
		return
	}

	address, err := a.app.CreateUserAddress(addr, uid)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, address)
}

func (a *API) getUserAddress(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	addrID, e := strconv.ParseInt(chi.URLParam(r, "address_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("getUserAddress", model.ErrInternal, locale.GetUserLocalizer("en"), msgURLParamErr, http.StatusInternalServerError, nil))
		return
	}

	address, err := a.app.GetUserAddress(uid, addrID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, address)
}

func (a *API) getUserAddresses(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	address, err := a.app.GetUserAddresses(uid)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, address)
}

func (a *API) updateUserAddress(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	addrID, e := strconv.ParseInt(chi.URLParam(r, "address_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("updateUserAddress", model.ErrInternal, locale.GetUserLocalizer("en"), msgURLParamErr, http.StatusInternalServerError, nil))
		return
	}

	patch, err := model.AddressPatchFromJSON(r.Body)
	if err != nil {
		respondError(w, model.NewAppErr("updateUserAddress", model.ErrInternal, locale.GetUserLocalizer("en"), msgAddressPatchFromJSON, http.StatusInternalServerError, nil))
		return
	}

	address, pErr := a.app.PatchUserAddress(uid, addrID, patch)
	if pErr != nil {
		respondError(w, pErr)
		return
	}
	respondJSON(w, http.StatusOK, address)
}

func (a *API) deleteUserAddress(w http.ResponseWriter, r *http.Request) {
	addrID, e := strconv.ParseInt(chi.URLParam(r, "address_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("deleteUserAddress", model.ErrInternal, locale.GetUserLocalizer("en"), msgDeleteUserAddress, http.StatusInternalServerError, nil))
		return
	}

	if err := a.app.DeleteUserAddress(addrID); err != nil {
		respondError(w, err)
		return
	}
	respondOK(w)
}

func (a *API) getUserOrders(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	userID, e := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("getUserOrders", model.ErrInternal, locale.GetUserLocalizer("en"), msgGetUserOrders, http.StatusInternalServerError, nil))
		return
	}

	if uid != userID {
		respondError(w, model.NewAppErr("getUserOrders", model.ErrInternal, locale.GetUserLocalizer("en"), msgGetUserOrders, http.StatusInternalServerError, nil))
		return
	}

	pages := pagination.NewFromRequest(r)
	orders, err := a.app.GetOrdersForUser(userID, pages.Limit(), pages.Offset())
	if err != nil {
		respondError(w, err)
		return
	}

	totalCount := -1
	if len(orders) > 0 {
		totalCount = orders[0].TotalCount
	}
	pages.SetData(orders, totalCount)

	respondJSON(w, http.StatusOK, pages)
}

func (a *API) createWishlist(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	props := model.MapStrInterfaceFromJSON(r.Body)
	productID, ok := props["product_id"].(float64)
	if !ok {
		respondError(w, model.NewAppErr("getUserOrders", model.ErrInternal, locale.GetUserLocalizer("en"), msgWishlistParamErr, http.StatusInternalServerError, nil))
		return
	}
	pid := int64(productID)

	err := a.app.CreateWishlistForUser(uid, pid)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{"status": "success"})
}

func (a *API) getWishlist(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	wishlist, err := a.app.GetWishlistForUser(uid)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

func (a *API) deleteWishlist(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	pid, e := strconv.ParseInt(chi.URLParam(r, "product_id"), 10, 64)
	if e != nil {
		respondError(w, model.NewAppErr("deleteWishlist", model.ErrInternal, locale.GetUserLocalizer("en"), msgWishlistParamErr, http.StatusInternalServerError, nil))
		return
	}

	err := a.app.DeleteWishlistForUser(uid, pid)
	if err != nil {
		respondError(w, err)
		return
	}

	respondOK(w)
}

func (a *API) clearWishlist(w http.ResponseWriter, r *http.Request) {
	uid := a.app.GetUserIDFromContext(r.Context())
	err := a.app.ClearWishlistForUser(uid)
	if err != nil {
		respondError(w, err)
		return
	}

	respondOK(w)
}
