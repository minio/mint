import logging

def create_logger(logFile):
    """
    Creates a logging object and returns it
    """
    logger = logging.getLogger("miniopy_logger")
    logger.setLevel(logging.INFO)
 
    # create the logging file handler
    fh = logging.FileHandler(logFile,mode="w")
 
    fmt = '%(levelname)s: - %(message)s at LineNo:%(lineno)d of:%(funcName)s in:%(module)s'
    formatter = logging.Formatter(fmt)
    fh.setFormatter(formatter)
 
    # add handler to logger object
    logger.addHandler(fh)
    return logger


