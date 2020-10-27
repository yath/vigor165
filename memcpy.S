#include <regdef.h>

.text
.globl my_memcpy
.set		noreorder
.ent my_memcpy
my_memcpy:
  or      v1,a0,a1
  andi    v1,v1,0x3
  bnez    v1,LAB_80024524
  move    v0,a0
  sltiu   v1,a2,4
  bnez    v1,LAB_80024524
  move    t0,a2
  move    a3,a1
  move    v1,a0
  addiu   a3,a3,4
LAB_800244e4:
  lw      t1,-4(a3)
  addiu   v1,v1,4
  addiu   t0,t0,-4
  sw      t1,-4(v1)
  sltiu   t1,t0,4
  beqz    t1,LAB_800244e4
  addiu   a3,a3,4
  addiu   a3,a3,-4
  addiu   v1,a2,-4
  li      a0,-4
  andi    a2,v1,0x3
  and     v1,v1,a0
  addiu   v1,v1,4
  addu    a1,a1,v1
  b       LAB_80024528
  addu    v1,v0,v1
LAB_80024524:
  move    v1,v0
LAB_80024528:
  beqz    a2,LAB_80024544
  addu    a3,v1,a2
LAB_80024530:
  addiu   a1,a1,1
  lb      a2,-1(a1)
  addiu   v1,v1,1
  bne     a3,v1,LAB_80024530
  sb      a2,-1(v1)
LAB_80024544:  
  jr      ra
nop
.end my_memcpy
