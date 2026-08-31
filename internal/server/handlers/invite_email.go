package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"GoTodo/internal/mailer"
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
)

// siteInviteRegisterURL is the signup link included in site invite emails.
func siteInviteRegisterURL(r *http.Request, email, token string) string {
	q := url.Values{}
	q.Set("email", email)
	q.Set("invite", token)
	return utils.AbsoluteURLForRequest(r, "/register?"+q.Encode())
}

// emailSiteInvite emails a signup link with the invite token prefilled.
// Send failures are ignored so the invite row still exists if mail is unconfigured.
func emailSiteInvite(r *http.Request, email, token string) {
	siteName := "GoTodo"
	settings, err := storage.GetSiteSettings()
	if err == nil && settings != nil && strings.TrimSpace(settings.SiteName) != "" {
		siteName = settings.SiteName
	}
	registerURL := siteInviteRegisterURL(r, email, token)
	subject := fmt.Sprintf("%s - You're invited", siteName)
	body := fmt.Sprintf(`Hello,

You're invited to join %s. Create your account here:

%s

If the link does not work, enter your email and the following invite token on the registration page:
%s

If you did not expect this invitation, you can ignore this email.
`, siteName, registerURL, token)
	_ = mailer.SendEmail(storage.SiteEmailConfig(settings), mailer.TriggerSiteInvite, subject, body, email)
}
