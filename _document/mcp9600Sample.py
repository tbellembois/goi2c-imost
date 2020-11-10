class MCP9600:
    def __init__(self, i2c_addr=I2C_ADDRESS_DEFAULT, i2c_dev=None):
        self._is_setup = False
        self._i2c_addr = i2c_addr
        self._i2c_dev = i2c_dev
        self._mcp9600 = Device(I2C_ADDRESSES, i2c_dev=self._i2c_dev, bit_width=8, registers=(
            Register('HOT_JUNCTION', 0x00, fields=(
                BitField('temperature', 0xFFFF, adapter=TemperatureAdapter()),
            ), bit_width=16),
            Register('DELTA', 0x01, fields=(
                BitField('value', 0xFFFF, adapter=TemperatureAdapter()),
            ), bit_width=16),
            Register('COLD_JUNCTION', 0x02, fields=(
                BitField('temperature', 0x1FFF, adapter=TemperatureAdapter()),
            ), bit_width=16),
            Register('RAW_DATA', 0x03, fields=(
                BitField('adc', 0xFFFFFF),
            ), bit_width=24),
            Register('STATUS', 0x04, fields=(
                BitField('burst_complete', 0b10000000),
                BitField('updated', 0b01000000),
                BitField('input_range', 0b00010000),
                BitField('alert_4', 0b00001000),
                BitField('alert_3', 0b00000100),
                BitField('alert_2', 0b00000010),
                BitField('alert_1', 0b00000001)
            )),
            Register('THERMOCOUPLE_CONFIG', 0x05, fields=(
                BitField('type_select', 0b01110000, adapter=LookupAdapter({
                    'K': 0b000,
                    'J': 0b001,
                    'T': 0b010,
                    'N': 0b011,
                    'S': 0b100,
                    'E': 0b101,
                    'B': 0b110,
                    'R': 0b111
                })),
                BitField('filter_coefficients', 0b00000111)
            )),
            Register('DEVICE_CONFIG', 0x06, fields=(
                BitField('cold_junction_resolution', 0b10000000, adapter=LookupAdapter({
                    0.0625: 0b0,
                    0.25: 0b1
                })),
                BitField('adc_resolution', 0b01100000, adapter=LookupAdapter({
                    18: 0b00,
                    16: 0b01,
                    14: 0b10,
                    12: 0b11
                })),
                BitField('burst_mode_samples', 0b00011100, adapter=LookupAdapter({
                    1: 0b000,
                    2: 0b001,
                    4: 0b010,
                    8: 0b011,
                    16: 0b100,
                    32: 0b101,
                    64: 0b110,
                    128: 0b111
                })),
                BitField('shutdown_modes', 0b00000011, adapter=LookupAdapter({
                    'Normal': 0b00,
                    'Shutdown': 0b01,
                    'Burst': 0b10
                }))
            )),
            Register('ALERT1_CONFIG', 0x08, fields=(
                BitField('clear_interrupt', 0b10000000),
                BitField('monitor_junction', 0b00010000),  # 1 Cold Junction, 0 Thermocouple
                BitField('rise_fall', 0b00001000),         # 1 rising, 0 cooling
                BitField('state', 0b00000100),             # 1 active high, 0 active low
                BitField('mode', 0b00000010),              # 1 interrupt mode, 0 comparator mode
                BitField('enable', 0b00000001)             # 1 enable, 0 disable
            )),
            Register('ALERT2_CONFIG', 0x09, fields=(
                BitField('clear_interrupt', 0b10000000),
                BitField('monitor_junction', 0b00010000),  # 1 Cold Junction, 0 Thermocouple
                BitField('rise_fall', 0b00001000),         # 1 rising, 0 cooling
                BitField('state', 0b00000100),             # 1 active high, 0 active low
                BitField('mode', 0b00000010),              # 1 interrupt mode, 0 comparator mode
                BitField('enable', 0b00000001)             # 1 enable, 0 disable
            )),
            Register('ALERT3_CONFIG', 0x0A, fields=(
                BitField('clear_interrupt', 0b10000000),
                BitField('monitor_junction', 0b00010000),  # 1 Cold Junction, 0 Thermocouple
                BitField('rise_fall', 0b00001000),         # 1 rising, 0 cooling
                BitField('state', 0b00000100),             # 1 active high, 0 active low
                BitField('mode', 0b00000010),              # 1 interrupt mode, 0 comparator mode
                BitField('enable', 0b00000001)             # 1 enable, 0 disable
            )),
            Register('ALERT4_CONFIG', 0x0B, fields=(
                BitField('clear_interrupt', 0b10000000),
                BitField('monitor_junction', 0b00010000),  # 1 Cold Junction, 0 Thermocouple
                BitField('rise_fall', 0b00001000),         # 1 rising, 0 cooling
                BitField('state', 0b00000100),             # 1 active high, 0 active low
                BitField('mode', 0b00000010),              # 1 interrupt mode, 0 comparator mode
                BitField('enable', 0b00000001)             # 1 enable, 0 disable
            )),
            Register('ALERT1_HYSTERESIS', 0x0C, fields=(
                BitField('value', 0xFF),
            )),
            Register('ALERT2_HYSTERESIS', 0x0D, fields=(
                BitField('value', 0xFF),
            )),
            Register('ALERT3_HYSTERESIS', 0x0E, fields=(
                BitField('value', 0xFF),
            )),
            Register('ALERT4_HYSTERESIS', 0x0F, fields=(
                BitField('value', 0xFF),
            )),
            Register('ALERT1_LIMIT', 0x10, fields=(
                BitField('value', 0xFFFF, adapter=AlertLimitAdapter()),
            ), bit_width=16),
            Register('ALERT2_LIMIT', 0x11, fields=(
                BitField('value', 0xFFFF, adapter=AlertLimitAdapter()),
            ), bit_width=16),
            Register('ALERT3_LIMIT', 0x12, fields=(
                BitField('value', 0xFFFF, adapter=AlertLimitAdapter()),
            ), bit_width=16),
            Register('ALERT4_LIMIT', 0x13, fields=(
                BitField('value', 0xFFFF, adapter=AlertLimitAdapter()),
            ), bit_width=16),
            Register('CHIP_ID', 0x20, fields=(
                BitField('id', 0xFF00),
                BitField('revision', 0x00FF, adapter=RevisionAdapter())
            ), bit_width=16)
        ))