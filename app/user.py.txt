from flask_login import UserMixin
from flask_sqlalchemy import SQLAlchemy
from serializers import AlchemyEncoder
import pandas as pd
import sys
import os

# configuration used to connect to MariaDB
# config = {
#     "host": os.getenv("DBHOST", "127.0.0.1"),
#     "port": int(os.getenv("DBPORT", "3306")),
#     "user": os.getenv("DBUSER", "root"),
#     "password": os.getenv("DBPASS", "password"),
#     "database": os.getenv("DBNAME", "konsumo"),
# }


def get_db():
    # try:
    #     conn = mariadb.session.connect(**config)
    # except mariadb.session.Error as e:
    #     print(f"Error connecting to MariaDB Platform: {e}")
    #     sys.exit(1)
    conn = engine.connect()

    return conn

class User(UserMixin):
    
    def __init__(self, id_=None, name=None, email=None, profile_pic=None, location="",db=None):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic
        self.location = location

    def get(self, user_id):
        
        try:
            user = db.session.execute( f"SELECT user_id, name, email, profile_pic, location FROM user WHERE user_id = %d", (user_id, ) )
            user = list(user[0])
        except:
            return False
        
        user = User(
            id_=user[0], name=user[1], email=user[2], profile_pic=user[3], location=user[4]
        )
        return user

    def create(self, id_, name, email, profile_pic):
        db.session.execute(
            f"INSERT INTO user (user_id, name, email, profile_pic) "
             "VALUES (?, ?, ?, ?)",
             (id_, name, email, profile_pic),
        )
        db.session.commit()
        db.session.close()
    
    def set_location(self, user_id, location):
        db.session.execute( f"UPDATE user set location = %s WHERE user_id = %d", (location, user_id) )
        db.session.commit()
        db.session.close()

    def set_data(self, date, type, value1, value2, user_id):
        if len(value2) > 0 :
            db.session.execute( f"INSERT INTO user_data (date, type, value1, value2, user_id) VALUES (?, ?, ?, ?, ?)", (date, type, value1, value2, user_id, ) )
        db.session.execute( f"INSERT INTO user_data (date, type, value1, user_id) VALUES (?, ?, ?, ?)", (date, type, value1, user_id, ) )
        db.session.commit()
        db.session.close()
    
    def get_data(self, user_id):
        try:
            rows = db.session.execute( f"SELECT date, type, value1, value2 FROM user_data WHERE user_id = %d ORDER BY date", (user_id, ) )
            return rows
        except:
            return False

    def db_close(self):
        db.session.close()
