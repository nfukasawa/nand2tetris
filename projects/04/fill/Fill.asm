// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.
(ENTRY)
    // @pos = @SCREEN
    @SCREEN
    D=A
    @pos
    M=D

    // if @KBD > 0 then goto @FILL
    @KBD
    D=M
    @FILL
    D;JGT

(CLEAR)
    // @color=0x0000
    @color
    M=0

    // goto @LOOP
    @LOOP
    0;JMP

(FILL)
    // @color=0xFFFF
    @color
    M=-1

(LOOP)
    // *@pos = @color
    @color
    D=M
    @pos
    A=M
    M=D

    // @pos++
    @pos
    M=M+1

    // if 512*256/16 + @SCREEN - @pos > 0 then goto @LOOP
    @8192
    D=A
    @SCREEN
    D=D+A
    @pos
    D=D-M  
    @LOOP
    D;JGT

    // goto @ENTRY
    @ENTRY
    0;JMP