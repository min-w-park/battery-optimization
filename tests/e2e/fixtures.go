package e2e

import "time"

// Test Configuration

const (
	// Service URLs (default Docker Compose ports)
	AssetManagementURL = "http://localhost:8080"
	MarketDataURL      = "http://localhost:8081"
	TelemetryURL       = "http://localhost:8082"
	DeviceInterfaceURL = "http://localhost:8083"
	NATSURL            = "nats://localhost:4222"

	// Timeouts
	DefaultTimeout       = 5 * time.Second
	EventTimeout         = 3 * time.Second
	NoEventTimeout       = 1 * time.Second
	ServiceStartTimeout  = 30 * time.Second
	ServiceRetryDelay    = 1 * time.Second
	MaxRetries           = 30
	BatteryStateInterval = 1 * time.Second

	// Test Data Constants
	DefaultCapacity    = 200.0
	DefaultMaxPower    = 100.0
	DefaultRampRate    = 10.0
	DefaultEfficiency  = 0.95
	DefaultInterval    = 5
	DefaultTemperature = 25.0
	DefaultVoltage     = 400.0
)

// Price Scenarios (from arbitrage algorithm thresholds)

const (
	// Charging threshold: < $50/MWh
	VeryLowPrice = 20.0
	LowPrice     = 30.0
	EdgeLowPrice = 49.0
	ThresholdLow = 50.0

	// Mid-range (no arbitrage opportunity)
	MidPrice = 75.0

	// Discharging threshold: > $100/MWh
	ThresholdHigh = 100.0
	EdgeHighPrice = 101.0
	HighPrice     = 150.0
	VeryHighPrice = 200.0
)

// SoC Scenarios (from arbitrage algorithm)

const (
	// Charging: SoC < 80%
	VeryLowSoC  = 10.0
	LowSoC      = 30.0
	MidSoC      = 50.0
	EdgeHighSoC = 79.0
	HighSoC     = 80.0
	VeryHighSoC = 95.0

	// Discharging: SoC > 30%
	MinSoCForDischarge   = 30.0
	TargetSoCCharging    = 80.0
	TargetSoCDischarging = 30.0
)

// Battery States

const (
	StateIdle        = "IDLE"
	StateCharging    = "CHARGING"
	StateDischarging = "DISCHARGING"
	StateError       = "ERROR"
	StateOffline     = "OFFLINE"
)

// Automation Modes

const (
	AutomationManual   = "MANUAL"
	AutomationSemiAuto = "SEMI_AUTO"
	AutomationFullAuto = "FULL_AUTO"
)

// Battery Fixtures

// DefaultBatterySpec returns standard battery specifications for testing
func DefaultBatterySpec() BatterySpec {
	return BatterySpec{
		Capacity:     DefaultCapacity,
		MaxPower:     DefaultMaxPower,
		RampRate:     DefaultRampRate,
		Efficiency:   DefaultEfficiency,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: BatteryConstraints{
			WarrantyEOL:    8000.0,
			MaxCycles:      10000,
			TempMin:        -20.0,
			TempMax:        60.0,
			GridCompliance: []string{"FCAS", "ENERGY"},
		},
	}
}

// LargeBatterySpec returns a larger battery for high-power testing
func LargeBatterySpec() BatterySpec {
	return BatterySpec{
		Capacity:     500.0,
		MaxPower:     250.0,
		RampRate:     25.0,
		Efficiency:   0.92,
		Location:     "NSW",
		Manufacturer: "BYD",
		Constraints: BatteryConstraints{
			WarrantyEOL:    10000.0,
			MaxCycles:      15000,
			TempMin:        -10.0,
			TempMax:        50.0,
			GridCompliance: []string{"FCAS", "ENERGY"},
		},
	}
}

// SmallBatterySpec returns a smaller battery for edge case testing
func SmallBatterySpec() BatterySpec {
	return BatterySpec{
		Capacity:     50.0,
		MaxPower:     25.0,
		RampRate:     5.0,
		Efficiency:   0.90,
		Location:     "VIC",
		Manufacturer: "Tesla",
		Constraints: BatteryConstraints{
			WarrantyEOL:    5000.0,
			MaxCycles:      8000,
			TempMin:        -15.0,
			TempMax:        55.0,
			GridCompliance: []string{"ENERGY"},
		},
	}
}

// BatterySpec represents battery specifications
type BatterySpec struct {
	Capacity     float64
	MaxPower     float64
	RampRate     float64
	Efficiency   float64
	Location     string
	Manufacturer string
	Constraints  BatteryConstraints
}

// BatteryConstraints represents battery operational constraints
type BatteryConstraints struct {
	WarrantyEOL    float64
	MaxCycles      int
	TempMin        float64
	TempMax        float64
	GridCompliance []string
}

// Battery State Fixtures

// IdleBatteryState returns a battery in idle state with mid SoC
func IdleBatteryState() BatteryState {
	return BatteryState{
		SoC:            MidSoC,
		Power:          0.0,
		OperationState: StateIdle,
		Temperature:    DefaultTemperature,
		Voltage:        DefaultVoltage,
		Current:        0.0,
	}
}

// LowSoCBatteryState returns a battery with low SoC (ready for charging)
func LowSoCBatteryState() BatteryState {
	return BatteryState{
		SoC:            LowSoC,
		Power:          0.0,
		OperationState: StateIdle,
		Temperature:    DefaultTemperature,
		Voltage:        DefaultVoltage,
		Current:        0.0,
	}
}

// HighSoCBatteryState returns a battery with high SoC (ready for discharging)
func HighSoCBatteryState() BatteryState {
	return BatteryState{
		SoC:            VeryHighSoC,
		Power:          0.0,
		OperationState: StateIdle,
		Temperature:    DefaultTemperature,
		Voltage:        DefaultVoltage,
		Current:        0.0,
	}
}

// ChargingBatteryState returns a battery currently charging
func ChargingBatteryState() BatteryState {
	return BatteryState{
		SoC:            MidSoC,
		Power:          50.0,
		OperationState: StateCharging,
		Temperature:    DefaultTemperature,
		Voltage:        DefaultVoltage,
		Current:        50.0 / DefaultVoltage,
	}
}

// DischargingBatteryState returns a battery currently discharging
func DischargingBatteryState() BatteryState {
	return BatteryState{
		SoC:            MidSoC,
		Power:          -50.0,
		OperationState: StateDischarging,
		Temperature:    DefaultTemperature,
		Voltage:        DefaultVoltage,
		Current:        -50.0 / DefaultVoltage,
	}
}

// BatteryState represents real-time battery state
type BatteryState struct {
	SoC            float64
	Power          float64
	OperationState string
	Temperature    float64
	Voltage        float64
	Current        float64
}

// Price Fixtures

// ChargingOpportunityPrice returns a price that triggers charging
func ChargingOpportunityPrice() float64 {
	return LowPrice
}

// DischargingOpportunityPrice returns a price that triggers discharging
func DischargingOpportunityPrice() float64 {
	return HighPrice
}

// NoOpportunityPrice returns a price that triggers no arbitrage
func NoOpportunityPrice() float64 {
	return MidPrice
}

// EdgeCaseChargingPrice returns a price just below the charging threshold
func EdgeCaseChargingPrice() float64 {
	return EdgeLowPrice
}

// EdgeCaseDischargingPrice returns a price just above the discharging threshold
func EdgeCaseDischargingPrice() float64 {
	return EdgeHighPrice
}

// Test Scenarios

// TestScenario represents a complete test scenario with expected outcomes
type TestScenario struct {
	Name                string
	BatterySpec         BatterySpec
	InitialState        BatteryState
	Price               float64
	ExpectedOpportunity string // "CHARGING", "DISCHARGING", "NONE"
	ExpectedCommand     bool
	Description         string
}

// GetChargingScenario returns a scenario that should trigger charging
func GetChargingScenario() TestScenario {
	return TestScenario{
		Name:                "Charging Opportunity",
		BatterySpec:         DefaultBatterySpec(),
		InitialState:        LowSoCBatteryState(),
		Price:               ChargingOpportunityPrice(),
		ExpectedOpportunity: "CHARGING",
		ExpectedCommand:     true,
		Description:         "Low price ($30/MWh) + Low SoC (30%) → Charging opportunity",
	}
}

// GetDischargingScenario returns a scenario that should trigger discharging
func GetDischargingScenario() TestScenario {
	return TestScenario{
		Name:                "Discharging Opportunity",
		BatterySpec:         DefaultBatterySpec(),
		InitialState:        HighSoCBatteryState(),
		Price:               DischargingOpportunityPrice(),
		ExpectedOpportunity: "DISCHARGING",
		ExpectedCommand:     true,
		Description:         "High price ($150/MWh) + High SoC (95%) → Discharging opportunity",
	}
}

// GetNoOpportunityScenario returns a scenario with no arbitrage opportunity
func GetNoOpportunityScenario() TestScenario {
	return TestScenario{
		Name:                "No Opportunity",
		BatterySpec:         DefaultBatterySpec(),
		InitialState:        IdleBatteryState(),
		Price:               NoOpportunityPrice(),
		ExpectedOpportunity: "NONE",
		ExpectedCommand:     false,
		Description:         "Mid price ($75/MWh) + Mid SoC (50%) → No opportunity",
	}
}

// GetEdgeCaseChargingScenario returns a scenario at the charging threshold edge
func GetEdgeCaseChargingScenario() TestScenario {
	return TestScenario{
		Name:                "Edge Case Charging",
		BatterySpec:         DefaultBatterySpec(),
		InitialState:        LowSoCBatteryState(),
		Price:               EdgeLowPrice,
		ExpectedOpportunity: "CHARGING",
		ExpectedCommand:     true,
		Description:         "Edge price ($49/MWh) + Low SoC (30%) → Should trigger charging",
	}
}

// GetEdgeCaseDischargingScenario returns a scenario at the discharging threshold edge
func GetEdgeCaseDischargingScenario() TestScenario {
	return TestScenario{
		Name:                "Edge Case Discharging",
		BatterySpec:         DefaultBatterySpec(),
		InitialState:        HighSoCBatteryState(),
		Price:               EdgeHighPrice,
		ExpectedOpportunity: "DISCHARGING",
		ExpectedCommand:     true,
		Description:         "Edge price ($101/MWh) + High SoC (95%) → Should trigger discharging",
	}
}
