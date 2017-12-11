# Changelog

## [0.2.1] - 2017-12-10

### Added
- Godocs for functions and types
- Separate examples
- Integration tests

## [0.2.0] - 2017-10-06

### Added
- More predefined OBD commands for getting:
  - Short term fuel trim bank 1
  - Long term fuel trim bank 1
  - Short term fuel trim bank 2
  - Long term fuel trim bank 2
  - OBD standards vehicle conforms to
  - Supported PIDs 21-40
  - Supported PIDs 41-60
  - Supported PIDs 61-80
  - Supported PIDs 81-A0

### Changed
- More clear handling of results and conversions

## [0.1.0] - 2017-09-24

### Added
- Connecting to the device
- Writing data to the device
- Reading data from the device
- Automatically setting end protocol
- Checking supported OBD commands
- Getting version of device
- Type safe OBD command usage API
- Predefined OBD sensor commands for getting:
  - Engine load
  - Coolant temperature
  - Fuel pressure
  - Intake manifold pressure
  - Engine RPM
  - Vehicle speed
  - Timing advance
  - MAF air flow rate
  - Throttle position
- Getting all supported OBD sensor data for the current car
- Exposing Prometheus metrics over HTTP

### Performance

Bench results:

```
BenchmarkParseSensorData4-4      2000000               955 ns/op
BenchmarkParseSensorData2-4      2000000               823 ns/op
BenchmarkBytesToUint8-4         100000000               17.6 ns/op
BenchmarkBytesToUint4-4         100000000               10.4 ns/op
BenchmarkHexLitsToBytes8-4      10000000               215 ns/op
BenchmarkHexLitsToBytes4-4      20000000               117 ns/op
BenchmarkBytesToBits8-4          3000000               523 ns/op
BenchmarkBytesToBits4-4          5000000               282 ns/op

Intel(R) Core(TM) i5 CPU         760  @ 2.80GHz
```

Reading sensor data from my Lexus IS200 -04 was fairly slow. At worst it took
almost ~2.5 seconds, and at best it took ~1.7 seconds:

**Slowest**
```
2017/09/09 19:46:28 -- Starting to read sensors
2017/09/09 19:46:31 -- engine_load: 0.337255
2017/09/09 19:46:31 -- coolant_temperature: 61
2017/09/09 19:46:31 -- intake_manifold_pressure: 33
2017/09/09 19:46:31 -- engine_rpm: 937.500000
2017/09/09 19:46:31 -- vehicle_speed: 0
2017/09/09 19:46:31 -- timing_advance: 11.000000
2017/09/09 19:46:31 -- throttle_position: 0.156863
2017/09/09 19:46:31 Took 2.430519661s to read
```

**Fastest**
```
2017/09/09 19:47:14 -- Starting to read sensors
2017/09/09 19:47:16 -- engine_load: 0.333333
2017/09/09 19:47:16 -- coolant_temperature: 63
2017/09/09 19:47:16 -- intake_manifold_pressure: 33
2017/09/09 19:47:16 -- engine_rpm: 939.750000
2017/09/09 19:47:16 -- vehicle_speed: 0
2017/09/09 19:47:16 -- timing_advance: 11.000000
2017/09/09 19:47:16 -- throttle_position: 0.164706
2017/09/09 19:47:16 Took 1.734316157s to read
```

This could probably be sped up if we tell the ELM327 device how many data lines
to receive, instead of waiting the default 200 ms per OBD command. See ELM327
data sheet for details, search for "data lines", or similar.
