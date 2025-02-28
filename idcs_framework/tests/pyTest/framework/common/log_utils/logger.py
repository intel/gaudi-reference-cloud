import logging
import sys
import os

LOGS_DIR = "logs"
LOG_FILE_NAME = "test_execution.log"


def get_logger(name=None, request=None):
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)
    # Create handler for Report Portal if the service has been
    # configured and started.
    if request and hasattr(request.node.config, 'py_test_service'):

        # Add additional handlers if it is necessary
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(logging.INFO)
        logger.addHandler(console_handler)
    else:
        formatter = logging.Formatter(
            "%(asctime)s | %(levelname)s | %(filename)s:%(lineno)s | "
            "%(funcName)s | %(message)s"
        )
        file_path = os.path.join(str(os.getcwd()), LOGS_DIR, LOG_FILE_NAME)

        directory = os.path.dirname(file_path)
        if not os.path.exists(directory):
            os.makedirs(directory)
        file_handler = logging.FileHandler(file_path)
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)
    return logger
