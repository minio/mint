import functools
import inspect
def log_decorate(logger):
    """
    A decorator that wraps the passed in function and logs 
    exceptions should one occur
 
    @param logger: The logging object
    """
    def decorator(func):
 
        def wrapper(*args, **kwargs):
            func_name=inspect.stack()[1][4][0]
            file_name=inspect.stack()[1][1]
            line_no = inspect.stack()[1][2]
            print(file_name, ":", line_no, ":", func_name)
            logger.info("Entering " + func_name)
            try:
                rc = func(*args, **kwargs)
                logger.info("Exiting " + func_name)
                return rc
            except:
                # log the exception
                err = "There was an exception in  "
                err += func.__name__
                logger.exception(err)
 
            # re-raise the exception
            raise
        return wrapper
    return decorator