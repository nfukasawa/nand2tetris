defmodule HackAsmTest do
  use ExUnit.Case
  doctest HackAsm

  test "greets the world" do
    assert HackAsm.hello() == :world
  end
end
