from flask import g, current_app
from flask_login import UserMixin
from flask_sqlalchemy import SQLAlchemy
from konsumo import db
import enum

DEBUG = True

# def get_db():
#   if 'db' not in g:
#     g.db = SQLAlchemy(current_app)
#   return g.db

class User(db.Model, UserMixin):
    __tablename__ = "user"
    user_id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(125))
    email = db.Column(db.String(250))
    profile_pic = db.Column(db.String(255))
    location = db.Column(db.String(200))

    def __init__(self, id_=None, name=None, email=None, profile_pic=None, location=""):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic
        self.location = location

    def get(self, user_id):
        sql = db.select(
                    User.user_id,User.name,User.email,User.profile_pic,User.location
                ).where(User.user_id == user_id)
        
        try:
            row = db.session.execute(sql) 
            user = list(row.fetchall())[0]
        except Exception as e:
            if DEBUG:
                print(e)
            return False
        
        user = User(
            id_=user[0], name=user[1], email=user[2], profile_pic=user[3], location=user[4]
        )
        return user

    def create(self,user_id, name, email, profile_pic):
        sql = db.insert(User).values(user_id=user_id, name=name, email=email, profile_pic=profile_pic)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()
    
    def set_location(self, user_id, location):
        sql = db.update(User).where(user_id == user_id).values(location=location)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()

    def set_data(self, date, type, value1, value2, user_id):
        if len(value2) > 0 :
            sql = db.insert(User).values(date=date, type=type, value1=value1, value2=value2, user_id=user_id)
        else :
            sql = db.insert(User).values(date=date, type=type, value1=value1, user_id=user_id)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()
    
    def get_data(self, user_id, type, value2=False):
        if value2:
            sql = db.select(
                    User_Data.date, User_Data.value1, User_Data.value2
                ).order_by(User_Data.date).where(User_Data.type == type).where(User_Data.user_id == user_id)
        else:
            sql = db.select(
                    User_Data.date, User_Data.value1
                ).order_by(User_Data.date).where(User_Data.type == type).where(User_Data.user_id == user_id)
        
        try:
            rows = db.session.execute(sql)
            return rows.fetchall()
        except Exception as e:
            if DEBUG:
                print(e)
            return False

    def db_close(self):
        db.session.close()

class TypeEnum(enum.Enum):
    gazoline=1
    water=2
    electricity=3
    other=4

class User_Data(db.Model):
    __tablename__ = "user_data"
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    date = db.Column(db.Date)
    type = db.Column(db.Enum(TypeEnum))
    value1 = db.Column(db.Integer)
    value2 = db.Column(db.Integer, nullable=True)
    user_id = db.Column(db.ForeignKey("user.user_id"), nullable=False)

# def init_db():
#     db = get_db()
#     db.create_all()

# if __name__ == "__main__":
#     init_db()