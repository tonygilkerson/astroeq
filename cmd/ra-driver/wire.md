
| Pico                         | Encoder                | TMC2208 Stepper Driver                | Nima17 Motor | UART0 Terminal   | UART1 Terminal   |
| ---------------------------- | ---------------------- | ------------------------------------- | ------------ | ---------------- | ---------------- |
| **GP0** - `UART0_TX_PIN`     |                        |                                       |              | **Pin1** - `TX`  |                  |
| **GP1** - `UART0_RX_PIN`     |                        |                                       |              | **Pin2** - `RX`  |                  |
| **GND**                      |                        |                                       |              | **Pin3** - `GND` |                  |
| GP2                          |                        |                                       |              |                  |                  |
| GP3                          |                        |                                       |              |                  |                  |
| **GP4** - `UART1 TX`         |                        |                                       |              |                  | **Pin1** - `TX`  |
| **GP5** - `UART1 RX`         |                        |                                       |              |                  | **Pin2** - `RX`  |
| GND                          |                        |                                       |              |                  | **Pin3** - `GND` |
| GP6                          |                        |                                       |              |                  |                  |
| GP7                          |                        |                                       |              |                  |                  |
| GP8                          |                        | **Pin16** - `DIR`                     |              |                  |                  |
| GP9                          |                        | **Pin15** - `STEP`                    |              |                  |                  |
| GND                          |                        |                                       |              |                  |                  |
| GP10                         |                        |                                       |              |                  |                  |
| GP11                         |                        | **Pin11** - `MS2`                     |              |                  |                  |
| GP12                         |                        | **Pin10** - `MS1`                     |              |                  |                  |
| GP13                         |                        | **Pin9** - `ENABLE` enabled=low       |              |                  |                  |
| GND                          |                        |                                       |              |                  |                  |
| GP14                         |                        |                                       |              |                  |                  |
| GP15                         |                        |                                       |              |                  |                  |
| SIDE END                     |                        |                                       |              |                  |                  |
| **VSYS**                     | **Pin1** - `VCC RED`   | **Pin2** - `VIO`                      |              |                  |                  |
| VSS                          |                        |                                       |              |                  |                  |
| GND                          |                        | **Pin1** - `GND`                      |              |                  |                  |
| 3v3                          |                        |                                       |              |                  |                  |
| 3v3(out)                     |                        |                                       |              |                  |                  |
| ADC_VREF                     |                        |                                       |              |                  |                  |
| GP28                         |                        |                                       |              |                  |                  |
| GND                          |                        | **Pin7** - `GND`                      |              |                  |                  |
| GP27                         |                        |                                       |              |                  |                  |
| GP26                         |                        |                                       |              |                  |                  |
| **RUN** - push button to GND |                        |                                       |              |                  |                  |
| GP22                         |                        |                                       |              |                  |                  |
| GND                          |                        |                                       |              |                  |                  |
| GP21                         |                        |                                       |              |                  |                  |
| **GP20** - `SPI0 CS`         | **Pin6** - `CS YEL`    |                                       |              |                  |                  |
| **GP19** - `SPI0_SDO_PIN`    | **Pin3** - `MOSI ORN`  |                                       |              |                  |                  |
| **GP18** - `SPI0_SCK_PIN`    | **Pin2** - `SCLK BRN`  |                                       |              |                  |                  |
| **GND**                      | **Pin4** - `GND  BLK`  |                                       |              |                  |                  |
| GP17                         |                        |                                       |              |                  |                  |
| **GP16** - `SPI0_SDI_PIN`    | **Pin5**  - `MISO GRN` |                                       |              |                  |                  |
|                              |                        | **Pin3** - `1B`                       | **Coil-1B**  |                  |                  |
|                              |                        | **Pin4** - `1A`                       | **Coil-1A**  |                  |                  |
|                              |                        | **Pin5** - `2A`                       | **Coil-2A**  |                  |                  |
|                              |                        | **Pin6** - `2B`                       | **Coil-2B**  |                  |                  |
|                              |                        | **Pin8** - `VMOT`  (12v power supply) |              |                  |                  |
|                              |                        |                                       |              |                  |                  |
| Pico                         | Encoder                | TMC2208 Stepper Driver                | Nima17 Motor | UART0 Terminal   | UART1 Terminal   |
|                              |                        |                                       |              |                  |                  |
|                              |