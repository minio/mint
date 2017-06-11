import logging

def create_logger(logFile):
    """
    Creates a logging object and returns it
    """
    logger = logging.getLogger("miniopy_logger")
    logger.setLevel(logging.INFO)
    # define a Handler which writes INFO messages or higher to the sys.stderr
    console = logging.StreamHandler()
    console.setLevel(logging.ERROR)
    # set a format which is simpler for console use
    formatter = logging.Formatter('%(name)-12s: %(levelname)-8s %(message)s')
    # tell the handler to use this format
    console.setFormatter(formatter)
    # add the handler to the root logger
    logging.getLogger('').addHandler(console)
    # create the logging file handler
    fh = logging.FileHandler(logFile,mode="w")
 
    fmt = '%(levelname)s: - %(message)s at LineNo:%(lineno)d of:%(funcName)s in:%(module)s'
    formatter = logging.Formatter(fmt)
    fh.setFormatter(formatter)
 
    # add handler to logger object
    logger.addHandler(fh)
    return logger


