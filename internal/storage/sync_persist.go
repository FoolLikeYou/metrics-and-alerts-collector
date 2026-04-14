package storage

import "log"

// SyncPersistMemStorage оборачивает MemStorage и после каждой мутации сохраняет снимок на диск.
// Используется при STORE_INTERVAL = 0.
type SyncPersistMemStorage struct {
	*MemStorage
	path string
}

// NewSyncPersistMemStorage возвращает обёртку; save при ошибке пишет в лог.
func NewSyncPersistMemStorage(inner *MemStorage, path string) *SyncPersistMemStorage {
	return &SyncPersistMemStorage{MemStorage: inner, path: path}
}

func (w *SyncPersistMemStorage) persist() {
	if err := SaveToJSONFile(w.path, w.MemStorage); err != nil {
		log.Printf("metrics persist: %v", err)
	}
}

func (w *SyncPersistMemStorage) SetGauge(name string, value float64) {
	w.MemStorage.SetGauge(name, value)
	w.persist()
}

func (w *SyncPersistMemStorage) AddCounter(name string, delta int64) {
	w.MemStorage.AddCounter(name, delta)
	w.persist()
}
