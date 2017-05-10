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

### Connections

 * PWMA = rpi GPIO21
 * AIN2 = 3.3V 
 * AIN1 = GND
 * STBY = rpi GPIO20
 * BIN1 = NC
 * BIN2 = NC
 * PWMB = NC
 * GND  = NC

 * VM = 7.2V
 * Vcc = 3.3V
 * GND = GND
 * A01 
 * A02 
 * B02 
 * B01
 * GND = GND

 


## RPi Hookup
 * /dev/serial0 for XV11 serial data
 * pwm21 for motor speed control
 * /dev/pi-blaster for pwm control
 * https://github.com/sarfata/pi-blaster
