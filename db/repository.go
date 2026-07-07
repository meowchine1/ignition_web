package db

type Repository interface {
	// Firmware
	AddFirmware(FileRecord) error
	DeleteFirmware(id int64) error
	ListFirmwares() ([]FileRecord, error)

	// Flasher
	AddFlasher(FileRecord) error
	DeleteFlasher(id int64) error
	ListFlashers() ([]FileRecord, error)
}