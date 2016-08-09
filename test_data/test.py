#!/usr/bin/python

def main():
  f = open('test_data.txt')
  x = f.readlines()
  f.close()
  x = [d.strip() for d in x]
  x = [int(d,16) for d in x]

  s = []
  for d in x:
    if d == 0xfa:
      print ','.join(s)
      s = []
    s.append('0x%02x' % d)

main()


