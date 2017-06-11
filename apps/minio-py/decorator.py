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
            try:
                print(func_name + "ENTRY")
                rc = func(*args, **kwargs)
                print(func_name + "EXIT")
                return rc
            except Exception as e:
                # log the exception
                err = "There was an error in  "
                err += func_name
                err += " at line no:" + str(line_no)
                err += " of :" + file_name + " message: " + e.message
                logger.info(err)

        return wrapper
    return decorator