package mock

import (
	crypto "crypto/rand"
	"fmt"
	"math/rand/v2"
	"metrika/internal/models"
	"sync/atomic"
	"time"
)

type Generator struct {
	idsCounter atomic.Int64
	intCount   atomic.Int64
}

func NewGenerator() *Generator {
	return &Generator{
		atomic.Int64{},
		atomic.Int64{},
	}
}

func (m *Generator) generateRandomUuid() string {
	var buf [16]byte
	//сначала записываем в слайс случайные байты
	_, err := crypto.Read(buf[:])
	//если генератор не работает - генерируем свой uuid
	if err != nil {
		t := time.Now().UnixMicro()
		c := m.idsCounter.Add(1)
		return fmt.Sprintf("%d-%d", t, c)
	}
	//преобразуем слайс байтов в шестнадцатеричную строку
	return fmt.Sprintf("%x", buf[:])
}


func (g *Generator) GenerateBucketSize(min, max int) int {
	return min + rand.IntN(max-min)
}

func (g *Generator) GenerateMockGuestSession(guest_id uint) *models.GuestSession {
	return &models.GuestSession{
		GuestID:    guest_id,
		Active:    true,
		IPAddress: g.generateRandomUuid(),
	}
}

func (g *Generator) GenerateMockEvent(session_id uint) *models.Event {
	return &models.Event{
		SessionID: session_id,
		Type:      "Generator",
		PageURL:   "Generator",
		Timestamp: time.Now(),
	}
}

func (g *Generator) GenerateMockGuest(domainId uint) models.Guest {
	return models.Guest{
		Fingerprint: g.generateRandomUuid(),
		DomainID:    domainId,
	}
}