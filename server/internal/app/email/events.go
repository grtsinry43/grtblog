package email

import appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"

var AvailableEventNames = buildAvailableEventNames()
var PublicSubscribableEventNames = buildPublicSubscribableEventNames()

func IsPublicSubscribableEventName(name string) bool {
	for _, item := range PublicSubscribableEventNames {
		if item == name {
			return true
		}
	}
	return false
}

func buildAvailableEventNames() []string {
	return appEvent.NamesByChannel(appEvent.ChannelEmail)
}

func buildPublicSubscribableEventNames() []string {
	items := EventCatalog()
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item.PublicEmail {
			result = append(result, item.Name)
		}
	}
	return result
}

func IsValidEventName(name string) bool {
	for _, item := range AvailableEventNames {
		if item == name {
			return true
		}
	}
	return false
}
