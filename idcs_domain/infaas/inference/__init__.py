import logging
import sys

logging.basicConfig(
    format="%(asctime)s %(levelname)s:%(filename)s:%(funcName)s: %(message)s",
    datefmt="%d-%m-%Y %H:%M:%S",
    level=logging.INFO,
    stream=sys.stdout
)
