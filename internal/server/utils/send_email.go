/*
**

	TODO: this is temporary solution to avoid circular dependency between storage and utils packages.
	This file should be removed once the email sending logic is moved to a separate package (e.g. mailer)
	and confirmed to work correctly with the storage package.

**
*/
package utils

import (
	"fmt"

	"GoTodo/internal/mailer"
	"GoTodo/internal/storage"
)

func SendEmail(subject, message, toEmail string) error {
	settings, err := storage.GetSiteSettings()
	if err != nil || settings == nil {
		return fmt.Errorf("email not configured")
	}
	return mailer.SendEmail(settings.MailerConfig(), subject, message, toEmail)
}
