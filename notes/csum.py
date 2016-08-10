#!/usr/bin/python


LIDAR_SOF_DELIMITER = 0xfa
LIDAR_MIN_INDEX = 0xa0
LIDAR_MAX_INDEX = 0xf9

LIDAR_START_OFS = 0
LIDAR_INDEX_OFS = 1
LIDAR_SPEED_OFS = 2
LIDAR_SAMPLE_OFS = 4
LIDAR_CHECKSUM_OFS = 20

def get_u16(s, ofs):
  return (s[ofs + 1] << 8) + s[ofs]

def checksum(s):
  cs = 0
  for i in range(0,LIDAR_CHECKSUM_OFS,2):
    cs = (cs << 1) + get_u16(s, i)
  return ((cs & 0x7fff) + (cs >> 15)) & 0x7fff

def valid(s):

  print ' '.join(['%02x' % c for c in s])

  if s[LIDAR_START_OFS] != LIDAR_SOF_DELIMITER:
    return False

  index = s[LIDAR_INDEX_OFS]
  if (index < LIDAR_MIN_INDEX) or (index > LIDAR_MAX_INDEX):
    return False

  cs = checksum(s)
  if cs != get_u16(s, LIDAR_CHECKSUM_OFS):
    return False

  return True

def main1():
  f = open('test1.txt')
  x = f.readlines()
  f.close()
  x = [l.strip() for l in x]
  x = [l.split() for l in x]
  samples = []
  for l in x:
    samples.append([int(b,16) for b in l])
  for s in samples:
    print ('bad', 'good')[valid(s)]

def process_frame(s):
  if len(s) != 22:
    return
  print ('bad', 'good')[valid(s)]

def main0():
  f = open('test0.txt')
  x = f.readlines()
  f.close()
  x = [d.strip() for d in x]
  x = [int(d,16) for d in x]

  s = []
  for d in x:
    if d == 0xfa:
      process_frame(s)
      s = []
    s.append(d)

main0()



