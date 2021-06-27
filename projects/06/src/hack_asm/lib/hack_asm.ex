defmodule HackAsm do
  def assemble(src, dest) do
    code = File.open!(src)

    try do
      {cmds, labels, vars} = IO.stream(code, :line) |> HackAsmParser.parse()
      code = HackAsmCode.generate(cmds, labels, vars) |> Enum.join("\n")
      File.write!(dest, code)
    after
      File.close(code)
    end
  end
end

defmodule HackAsmParser do
  def parse(raw_lines) do
    lines = raw_lines |> normalize_lines()

    {cmds, labels, vars, _} =
      lines
      |> Enum.reduce({[], %{}, %{}, 0}, fn line, {cmds, labels, vars, addr} ->
        {cmd, next_addr} = line |> parse_line(addr)

        next_cmds =
          case cmd do
            nil -> cmds
            %{type: :l} -> cmds
            cmd -> cmds ++ [Map.put(line, :cmd, cmd)]
          end

        next_labels =
          case cmd do
            %{type: :l, label: label, address: addr} -> Map.put(labels, label, addr)
            _ -> labels
          end

        next_vars =
          case cmd do
            %{type: :a, symbol: sym} ->
              if Map.has_key?(vars, sym) do
                vars
              else
                Map.put(vars, sym, Kernel.map_size(vars) + 0x0010)
              end

            _ ->
              vars
          end

        {next_cmds, next_labels, next_vars, next_addr}
      end)

    {cmds, labels, vars}
  end

  defp normalize_lines(raw_lines) do
    raw_lines
    |> Enum.with_index()
    |> Enum.reduce([], fn {line, index}, acc ->
      acc ++ [%{line: index, raw_code: line, code: normalize_line(line)}]
    end)
  end

  defp normalize_line(raw_line) do
    line =
      raw_line
      |> String.split(["//", "\n", "\r"])
      |> List.first()
      |> String.replace(" ", "")

    case line do
      "" -> nil
      line -> line
    end
  end

  # comment line
  defp parse_line(%{code: nil}, addr) do
    {nil, addr}
  end

  # parse label
  defp parse_line(%{code: "(" <> rest} = line, addr) do
    if rest |> String.last() == ")" do
      label =
        case rest |> String.trim_trailing(")") |> parse_symbol do
          :error -> raise "syntax error: line: #{line[:line]}, code: #{line[:raw_code]}"
          label -> label
        end

      {%{type: :l, label: label, address: addr}, addr}
    else
      raise "syntax error: line: #{line[:line]}, code: #{line[:raw_code]}"
    end
  end

  # parse A command
  defp parse_line(%{code: "@" <> value} = line, addr) do
    case parse_value(value) do
      :error ->
        case parse_symbol(value) do
          :error ->
            raise "syntax error: line: #{line[:line]}, code: #{line[:raw_code]}"

          sym ->
            {%{type: :a, symbol: sym}, addr + 1}
        end

      val ->
        {%{type: :a, value: val}, addr + 1}
    end
  end

  # parse C command
  defp parse_line(%{code: code}, addr) do
    cmd =
      case code |> String.split(";", parts: 2) do
        [dest_comp] ->
          [dest, comp] = dest_comp |> String.split("=", parts: 2)
          %{type: :c, comp: comp, dest: dest, jump: nil}

        [dest_comp, jump] ->
          case dest_comp |> String.split("=", parts: 2) do
            [comp] -> %{type: :c, comp: comp, dest: nil, jump: jump}
            [dest, comp] -> %{type: :c, comp: comp, dest: dest, jump: jump}
          end
      end

    {cmd, addr + 1}
  end

  defp parse_symbol(str) do
    if ~r/^[a-zA-Z.%_.$:][0-9a-zA-Z.%_.$:]*$/ |> Regex.match?(str) do
      str
    else
      :error
    end
  end

  defp parse_value(str) do
    case Integer.parse(str) do
      :error ->
        :error

      {i, _} ->
        if i < 0 do
          :error
        else
          i
        end
    end
  end
end

defmodule HackAsmCode do
  def generate(cmds, labels, vars) do
    cmds
    |> Enum.reduce([], fn %{cmd: %{type: ty}} = line, codes ->
      code =
        case ty do
          :a -> a_cmd(line, labels, vars)
          :c -> c_cmd(line)
        end

      codes ++ [code]
    end)
  end

  defp a_cmd(%{cmd: %{type: :a, value: val}}, _labels, _vars) do
    v = val |> Integer.to_string(2) |> String.pad_leading(15, "0")
    "0#{v}"
  end

  defp a_cmd(%{cmd: %{type: :a, symbol: sym}} = line, labels, vars) do
    val =
      case map_label(sym, labels) do
        nil ->
          case vars[sym] do
            nil ->
              raise "unknown label or variable #{sym}: line: #{line[:line]}, code: #{line[:raw_code]}"

            v ->
              v
          end

        v ->
          v
      end

    v = val |> Integer.to_string(2) |> String.pad_leading(15, "0")
    "0#{v}"
  end

  defp c_cmd(%{cmd: %{type: :c, comp: comp, dest: dest, jump: jump}} = line) do
    "111#{comp(comp, line)}#{dest(dest, line)}#{jump(jump, line)}"
  end

  defp comp(mnemonic, %{line: line, raw_code: raw_code}) do
    case mnemonic do
      "0" -> "0101010"
      "1" -> "0111111"
      "-1" -> "0111010"
      "D" -> "0001100"
      "A" -> "0110000"
      "!D" -> "0001101"
      "!A" -> "0110001"
      "-D" -> "0001111"
      "-A" -> "0110011"
      "D+1" -> "0011111"
      "A+1" -> "0110111"
      "D-1" -> "0001110"
      "A-1" -> "0110010"
      "D+A" -> "0000010"
      "D-A" -> "0010011"
      "A-D" -> "0000111"
      "D&A" -> "0000000"
      "D|A" -> "0010101"
      "M" -> "1110000"
      "!M" -> "1110001"
      "-M" -> "1110011"
      "M+1" -> "1110111"
      "M-1" -> "1110010"
      "D+M" -> "1000010"
      "D-M" -> "1010011"
      "M-D" -> "1000111"
      "D&M" -> "1000000"
      "D|M" -> "1010101"
      mnemonic -> raise "unknown comp mnemonic #{mnemonic}: line: #{line}, code: #{raw_code}"
    end
  end

  defp dest(mnemonic, %{line: line, raw_code: raw_code}) do
    case mnemonic do
      nil -> "000"
      "M" -> "001"
      "D" -> "010"
      "MD" -> "011"
      "A" -> "100"
      "AM" -> "101"
      "AD" -> "110"
      "AMD" -> "111"
      mnemonic -> raise "unknown dest mnemonic #{mnemonic}: line: #{line}, code: #{raw_code}"
    end
  end

  defp jump(mnemonic, %{line: line, raw_code: raw_code}) do
    case mnemonic do
      nil -> "000"
      "JGT" -> "001"
      "JEQ" -> "010"
      "JGE" -> "011"
      "JLT" -> "100"
      "JNE" -> "101"
      "JLE" -> "110"
      "JMP" -> "111"
      mnemonic -> raise "unknown jump mnemonic #{mnemonic}: line: #{line}, code: #{raw_code}"
    end
  end

  defp map_label(sym, labels) do
    case sym do
      "SP" ->
        0

      "LCL" ->
        1

      "ARG" ->
        2

      "THIS" ->
        3

      "THAT" ->
        4

      "R" <> n ->
        case Integer.parse(n) do
          {i, _} when i >= 0 and i <= 15 -> i
          _ -> nil
        end

      "SCREEN" ->
        16384

      "KBD" ->
        24576

      sym ->
        labels[sym]
    end
  end
end

defmodule HackAsm.CLI do
  def main(args) do
    {src, out} = args |> parse_args
    HackAsm.assemble(src, out)
  end

  defp parse_args(args) do
    {opts, args, _} =
      args
      |> OptionParser.parse(strict: [out: :string], aliases: [o: :out])

    out =
      case opts do
        [out: out] -> out
        _ -> usage()
      end

    src =
      case Enum.count(args) do
        1 -> List.first(args)
        _ -> usage()
      end

    {src, out}
  end

  defp usage() do
    IO.puts("usage: ./hack_asm -o <out> <source>")
    exit(1)
  end
end
