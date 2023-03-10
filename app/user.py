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
    
    def __init__(self, id_=None, name=None, email=None, profile_pic=None, location=""):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic
        self.location = location
        self.conn = get_db()
        self.db = self.conn.cursor()

    def get(self, user_id):
        self.db.execute( f"SELECT user_id, name, email, profile_pic, location FROM user WHERE user_id = %d", (user_id, ) )

        try:
            user = self.db.fetchall()
            user = list(user[0])
        except:
            return False
        
        user = User(
            id_=user[0], name=user[1], email=user[2], profile_pic=user[3], location=user[4]
        )
        # self.conn.close()
        return user

    def create(self, id_, name, email, profile_pic):
        self.db.execute(
            f"INSERT INTO user (user_id, name, email, profile_pic) "
             "VALUES (?, ?, ?, ?)",
             (id_, name, email, profile_pic),
        )
        self.conn.commit()
        self.conn.close()
    
    def set_location(self, user_id, location):
        self.db.execute( f"UPDATE user set location = %s WHERE user_id = %d", (location, user_id) )
        self.conn.commit()
        self.conn.close()

    def set_data(self, date, type, value1, value2, user_id):
        if len(value2) > 0 :
            self.db.execute( f"INSERT INTO user_data (date, type, value1, value2, user_id) VALUES (?, ?, ?, ?, ?)", (date, type, value1, value2, user_id, ) )
        self.db.execute( f"INSERT INTO user_data (date, type, value1, user_id) VALUES (?, ?, ?, ?)", (date, type, value1, user_id, ) )
        self.conn.commit()
        self.conn.close()

    def db_close(self):
        self.conn.close()