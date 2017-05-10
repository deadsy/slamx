# slamx

## System Power
 * DFRobot 25W DC-DC Power Module https://www.dfrobot.com/product-752.html
 * 7.2V In
 * 7.2V unregulated to XV11 motor
 * 5V to rpi

## XV11 Motor Driver
 * SparkFun ROB-09457 TB6612FNG Motor Driver
 * https://www.sparkfun.com/products/9457
 * Using Channel A

### Input Connections
 * PWMA = rpi gpio21 (used as pwm)
 * AIN2 = 3v3 
 * AIN1 = GND
 * STBY = rpi gpio20
 * BIN1 = NC
 * BIN2 = NC
 * PWMB = NC
 * GND  = NC

### Output Connections
 * VM = 7v2
 * Vcc = 3v3
 * GND = GND
 * A01 = Lidar Motor
 * A02 = Lidar Motor
 * B02 = NC 
 * B01 = NC
 * GND = GND

## RPi Hookup
 * /dev/serial0 for XV11 serial data
 * pwm21 for motor speed control
 * gpio20 for motor on/off
 * /dev/pi-blaster for pwm control
 * https://github.com/sarfata/pi-blaster
 * pi-blaster needs to have gpio20 added as a known pin
