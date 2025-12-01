package db

import (
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/templates"
	"pixpivot/arc/pkg/log"

	"gorm.io/gorm"
)

// SeedBreachNotificationTemplates seeds default DPDP-compliant templates
func SeedBreachNotificationTemplates(db *gorm.DB) {
	templates := []models.BreachNotificationTemplate{
		{
			TemplateName:     "dpb_notification_template",
			RecipientType:    "dpb",
			TemplateType:     "email",
			Subject:          "Data Breach Notification - {{breach_title}}",
			Body:             templates.DPBNotificationTemplate,
			Language:         "en",
			IsActive:         true,
			IsSystemTemplate: true,
		},
		{
			TemplateName:     "data_principal_notification_template",
			RecipientType:    "data_principal",
			TemplateType:     "email",
			Subject:          "Important Security Notice - Data Breach Notification",
			Body:             templates.DataPrincipalNotificationTemplate,
			Language:         "en",
			IsActive:         true,
			IsSystemTemplate: true,
		},
		{
			TemplateName:     "data_principal_notification_sms",
			RecipientType:    "data_principal",
			TemplateType:     "sms",
			Subject:          "",
			Body:             templates.DataPrincipalNotificationTemplateSMS,
			Language:         "en",
			IsActive:         true,
			IsSystemTemplate: true,
		},
		{
			TemplateName:     "dpb_followup_template",
			RecipientType:    "dpb",
			TemplateType:     "email",
			Subject:          "Follow-up on Data Breach Notification - {{breach_title}}",
			Body:             templates.DPBFollowUpTemplate,
			Language:         "en",
			IsActive:         true,
			IsSystemTemplate: true,
		},
	}

	for _, template := range templates {
		var existing models.BreachNotificationTemplate
		err := db.Where("template_name = ? AND recipient_type = ?", template.TemplateName, template.RecipientType).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&template).Error; err != nil {
				log.Logger.Error().Err(err).Str("template", template.TemplateName).Msg("Failed to seed breach notification template")
			} else {
				log.Logger.Info().Str("template", template.TemplateName).Msg("Seeded breach notification template")
			}
		}
	}
}
