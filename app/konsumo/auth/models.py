from flask_login import UserMixin
from konsumo import db
from random import choice
import uuid
import enum

DEBUG = True

def _generate_key():
    # Generate a secret key for API access
    secret_key=str(uuid.uuid4())+str(uuid.uuid4())
    return ''.join(choice((str.upper, str.lower))(c) for c in secret_key.replace('-',''))
class User(db.Model, UserMixin):
    __tablename__ = "user"
    user_id    = db.Column(db.VARCHAR(34), primary_key=True)
    name       = db.Column(db.String(125))
    email      = db.Column(db.String(250))
    profile_pic= db.Column(db.String(255))
    location   = db.Column(db.String(200))
    secret_key = db.Column(db.String(80))

    def __init__(self, id_=None, name=None, email=None, profile_pic=None):
        self.id = id_
        self.name = name
        self.email = email
        self.profile_pic = profile_pic

    @staticmethod
    def get(user_id):
        sql = db.select(
                    User.user_id,User.name,User.email,User.profile_pic
                ).where(User.user_id == user_id)
        
        try:
            rows = db.session.execute(sql)
            row = rows.all()[0]
        except Exception as e:
            if DEBUG:
                print(e)
            return False

        user = User(
            id_=row[0], name=row[1], email=row[2], profile_pic=row[3]
        )
        return user

    @staticmethod
    def get_info(user_id):
        sql = db.select(
                    User.user_id,User.secret_key,User.location
                ).where(User.user_id == user_id)
        
        rows = db.session.execute(sql)
        row = rows.all()[0]
        access_key = row[0]
        secret_key = row[1]
        location   = row[2]
        return location, access_key, secret_key

    @staticmethod
    def create(user_id, name, email, profile_pic):
        # Generate a secret key for API access
        secret_key=_generate_key()
        sql = db.insert(User).values(user_id=user_id, name=name, email=email, profile_pic=profile_pic, secret_key=secret_key)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()

    @staticmethod
    def regenerate_key(user_id):
        secret_key=_generate_key()
        sql = db.update(User).where(user_id == user_id).values(secret_key=secret_key)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()

    @staticmethod
    def set_location(user_id, location):
        sql = db.update(User).where(user_id == user_id).values(location=location)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()

    @staticmethod
    def set_data(date, type, value1, value2, user_id):
        if len(value2) > 0 :
            sql = db.insert(User_Data).values(date=date, type=type, value1=value1, value2=value2, user_id=user_id)
        else :
            sql = db.insert(User_Data).values(date=date, type=type, value1=value1, user_id=user_id)
        db.session.execute(sql)
        db.session.commit()
        db.session.close()
    
    @staticmethod
    def get_data(user_id, type, value2=False):
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
            return rows.all()
        except Exception as e:
            if DEBUG:
                print(e)
            return False

    @staticmethod
    def db_close():
        db.session.close()

class TypeEnum(enum.Enum):
    gazoline=1
    water=2
    electricity=3
    other=4

class User_Data(db.Model):
    __tablename__ = "user_data"

    class Meta:
        type_ = 'user_data'
        self_view_many = 'user_data'

    id = db.Column(db.BigInteger, primary_key=True, autoincrement=True)
    date = db.Column(db.Date, index=True)
    type = db.Column(db.Enum(TypeEnum), index=True)
    value1 = db.Column(db.Integer)
    value2 = db.Column(db.Integer, nullable=True)
    user_id = db.Column(db.ForeignKey("user.user_id"), nullable=False)
