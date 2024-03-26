package schedule_info

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetGroupScheduleInfo(
	startDateTime *time.Time,
	endDateTime *time.Time,
) *models.RequestSchedule {
	expiration := models.NewExpirationPattern()

	if endDateTime == nil {
		expiration.SetTypeEscaped(to.Ptr(models.NOEXPIRATION_EXPIRATIONPATTERNTYPE))
	} else {
		expiration.SetEndDateTime(endDateTime)
		expiration.SetTypeEscaped(to.Ptr(models.AFTERDATETIME_EXPIRATIONPATTERNTYPE))
	}

	requestSchedule := models.NewRequestSchedule()
	requestSchedule.SetExpiration(expiration)

	if startDateTime == nil {
		requestSchedule.SetStartDateTime(to.Ptr(time.Now()))
	} else {
		requestSchedule.SetStartDateTime(startDateTime)
	}

	return requestSchedule
}
