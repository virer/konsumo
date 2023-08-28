"""App entry point."""
from flask import redirect
from konsumo import create_app
import os
app = create_app()

DEBUG  = True
HOST   = os.getenv("HOST", "127.0.0.1")
PORT   = os.getenv("PORT", "8080")
SSL_CRT= os.getenv("SSL_CRT", "/ssl/cert.pem")
SSL_KEY= os.getenv("SSL_KEY", "/ssl/key.pem")
csrf = None

@app.route('/')
def root():
    return redirect("/konsumo", code=302)

if __name__ == "__main__":
    # SSL Mode
    app.run(host=HOST, port=int(PORT), ssl_context=(SSL_CRT, SSL_KEY), debug=DEBUG)