from collections import OrderedDict

from bs4 import BeautifulSoup


HEADER = """package cpu

// Generated from: http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html"""


def main():
    with open('opcodes.html', 'r') as f:
        soup = BeautifulSoup(f, 'html.parser')

    tables = soup.find_all('table')
    standard = generate(tables[0])
    prefix = generate(tables[1])

    print(HEADER)
    output('mnemonics', standard)
    print()
    output('prefixMnemonics', prefix)


def generate(table):
    mnemonics = OrderedDict()
    opcode = 0

    for row in table.find_all('tr')[1:]:
        for cell in row.find_all('td')[1:]:
            if len(cell.contents) > 1:
                mnemonics[opcode] = cell.contents[0]

            opcode += 1

    return mnemonics


def output(name, mnemonics):
    print('var', name, '= map[byte]string{')
    for opcode, mnemonic in mnemonics.items():
        print('\t0x{:02x}: "{}",'.format(opcode, mnemonic))
    print('}')


if __name__ == '__main__':
    main()
