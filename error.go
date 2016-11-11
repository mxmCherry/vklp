package vklp

import "fmt"

// "failed":1 — история событий устарела или была частично утеряна, приложение может получать события далее, используя новое значение ts из ответа.
// "failed":2 — истекло время действия ключа, необходимо заново получить key с помощью метода messages.getLongPollServer.
// "failed":3 — информация о пользователе утрачена, необходимо запросить новые key и ts с помощью метода messages.getLongPollServer.
// "failed": 4 — передан недопустимый номер версии в параметре version.

type Error uint8

func (e Error) Error() string {
	return fmt.Sprintf("vklp: error %d", e)
}
