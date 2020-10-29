# Vigor165 notes

This repository collects some notes on my reverse engineering efforts on the DrayTek Vigor 165.

The code in this repository contains a decompressor for the compressed firmware sections
starting with a 0x507f1daa (`0xaa 0x1d 0x7f 0x50`) header. The decompression routine has been
copied from the binary and running it in qemu works well enough that I didn’t bother
translating it into some HLL. It’s probably some LZO variant, but I haven’t found a matching
decompressor anywhere.

## Boot process

Some ROM (that also supports booting from UART and different boot configs, depending on the voltage
levels on some of the unpopulated jumper headers on the board) loads a U-Boot uImage from the SPI
flash at 0x6040 and executes it (don’t know much details). That code reads the DrayOS image from
flash at 0x10000, the length stored at the first word, to 0x86000000, verifies a checksum (really
only parity), copies 0x86000000+0x100 to 0x80024000 and jumps there. DrayTek has probably
inadvertedly published an old bootloader version with their [DrayTek 2760 GPL release](
http://gplsource.draytek.com/Vigor2760/v2760_v1213_GPL_release.tar.bz2), in
`build_dir/linux-ltqcpe_2_6_32_vrx288_gw_he_vdsl_nand/u-boot-2010.06/common/draycommon.c`.

The decompression routine has been copied from 0x80026d24, i.e. 0x10000+0x100+0x2d24 on the SPI
flash.

The running code looks like a Linux MIPS, although DrayTek claims it’s “DrayOS”, but I haven’t
looked into how the DTB relates to which binary is running on which processor. Are there two
independent Linux^WDrayOS kernels running? I don’t know.

## Undocumented features

* The secondary boot loader (the one on flash 0x6040) can be interrupted with `u` to drop to an
  U-Boot shell.  The regular code only uses the `sf probe` and `sf read` commands from U-Boot and
  then runs some own code to disable interrupts, so I’m not sure whether it’s at all possible to
  boot into DrayOS from that shell. `z` drops into a TFTP recovery mode.

* The UART main menu (“1: Enable TFTP Server”) has a bunch of undocumented commands. Most only
  toggle some debug level, but `m` allows for dumping memory. ☺

* The telnet/SSH CLI has a debug mode that can be entered with `sys admin drayteker`, unlocking
  more commands, e.g. `sys mem` for dumping memory. 0x60000000 is aliased to 0x80000000, as
  determined by trial and error and documented [here](https://patchwork.kernel.org/project/linux-clk/patch/20180803030237.3366-2-songjun.wu@linux.intel.com/)
  (in `arch/mips/include/asm/mach-intel-mips/kernel-entry-init.h`).

* The ftpd allows for retrieving `Router.bin` and `Router.web`, but this crashes the OS.

## Failures

* I haven’t found a (Linux) shell, e.g. busybox, nor any indication that DrayTek has left support
  for `fork()` in their Linux rip-off.

* The unpopulated second console header on the board didn’t seem to do anything, and I couldn’t
  get JTAG working on the “ICE” labeled unpopulated header. OpenOCD detects an IR len of 8, but
  it should be 5 for eJTAG, but I’m neither a JTAG nor a MIPS expert.
