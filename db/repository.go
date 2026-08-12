package db

type Repository interface {
	Close() error
	// Firmware
	AddFirmware(firmware FirmwareRecord) error
	DeleteFirmware(id int64) error
	UpdateFirmware(firmware FirmwareRecord) error
	GetFirmware(id int64) (*FirmwareRecord, error)
	FirmwareExists(filename string) (bool, error)
	
	ListAvailableFirmwares() ([]FirmwareRecord, error)
	ListFirmwares() ([]FirmwareRecord, error)
 
	EnableFirmware(id int64) error
	DisableFirmware(id int64) error
	ClearCurrentFirmware() error

	// Flasher
	AddFlasher(flasher FlasherRecord) error
	DeleteFlasher(id int64) error
	UpdateFlasher(flasher FlasherRecord) error
	GetFlasher(id int64) (*FlasherRecord, error)
	FlasherExists(filename string) (bool, error)

	ListFlashers() ([]FlasherRecord, error)
	ListFlashersByOS(os OSType) ([]FlasherRecord, error)

	GetCurrentFlasher(os OSType) (*FlasherRecord, error)
	SetCurrentFlasher(id int64, os OSType) error
	UnsetCurrentFlasher(os OSType) error

	// Statistics
	CountFirmwares() (int, error)
	CountFlashers() (int, error)

}