package db

type Repository interface {  
	// Firmware
	AddFirmware(FirmwareRecord) error
	DeleteFirmware(id int64) error
	UpdateFirmware(FirmwareRecord) error
	GetFirmware(id int64) (*FirmwareRecord, error)

	ListFirmwares() ([]FirmwareRecord, error)
	
	GetCurrentFirmware() (*FirmwareRecord, error)
	SetCurrentFirmware(id int64)(error)
	ClearCurrentFirmware() error
	 
	// Flasher
	AddFlasher(FlasherRecord) error
	DeleteFlasher(id int64) error
	UpdateFlasher(FirmwareRecord) error
	GetFlasher(id int64) (*FlasherRecord, error)

	ListFlashers() ([]FlasherRecord, error)

	GetCurrentFlasher(OSType) (*FlasherRecord, error)
	SetCurrentFlasher(id int64, OSType)(error)
	ClearCurrentFlasher(os OSType) error

	// =========================
	// Statistics
	// =========================
	CountFirmwares() (int, error)
	CountFlashers() (int, error)
	 
}