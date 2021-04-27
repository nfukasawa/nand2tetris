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
(LOOP)
    // if @KBD > 0 then goto FILL
    @KBD
    D=M
    @FILL
    D;JGT

(CLEAR)
    // n = 512*256/16
    @8192
    D=A

(CLEAR_LOOP)
    // n--
    D=D-1

    // *(@SCREEN+n)=0x0000
    @SCREEN
    A=A+D
    M=0
    
    // if n >= 0 then goto CLEAR_LOOP
    @CLEAR_LOOP
    D;JGE

    // go to LOOP
    @LOOP
    0;JMP
    
(FILL)
    // n = 512*256/16
    @8192
    D=A

(FILL_LOOP)
    // n--
    D=D-1

    // *(@SCREEN+n)=0xFFFF
    @SCREEN
    A=A+D
    M=-1

    // if n >= 0 then goto FILL_LOOP
    @FILL_LOOP
    D;JGE

    // go to LOOP
    @LOOP
    0;JMP