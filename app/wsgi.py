"""App entry point."""
from konsumo import create_app
import os
app = create_app()

DEBUG= True
HOST = os.environ.get("HOST", "127.0.0.1")
PORT = os.environ.get("PORT", "8080")

if __name__ == "__main__":
    # SSL Mode
    app.run(host=HOST, port=int(PORT), ssl_context="adhoc", debug=DEBUG)
    # No SSL (usage with gunicorn)
    # app.run(host=HOST, port=int(PORT), debug=DEBUG)