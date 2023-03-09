from flask_login import UserMixin
import mariadb
import sys
import os

# configuration used to connect to MariaDB
config = {
    "host": os.getenv("DBHOST", "127.0.0.1"),
    "port": int(os.getenv("DBPORT", "3306")),
    "user": os.getenv("DBUSER", "root"),
    "password": os.getenv("DBPASS", "password"),
    "database": os.getenv("DBNAME", "konsumo"),
}

def get_db():
    try:
        conn = mariadb.connect(**config)
    except mariadb.Error as e:
        print(f"Error connecting to MariaDB Platform: {e}")
        sys.exit(1)

    return conn

class User(UserMixin):
    
    def __init__(self, id_=None, name=None, email=None, profile_pic=None):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic
        self.conn = get_db()
        self.db = self.conn.cursor()

    def get(self, user_id):
        sql = "SELECT * FROM user WHERE id = {}".format(user_id)
        self.db.execute( sql )

        try:
            user = self.db.fetchall()
            user = list(user[0])
        except:
            return False
        
        user = User(
            id_=user[0], name=user[1], email=user[2], profile_pic=user[3]
        )
        self.conn.close()
        return user

    def create(self, id_, name, email, profile_pic):
        self.db.execute(
            "INSERT INTO user (id, name, email, profile_pic) "
            "VALUES (?, ?, ?, ?)",
            (id_, name, email, profile_pic),
        )
        self.conn.commit()
        self.conn.close()
