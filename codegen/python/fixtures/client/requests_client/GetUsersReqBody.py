"""
Auto-generated class for GetUsersReqBody
"""

from . import client_support


class GetUsersReqBody(object):
    """
    auto-generated. don't touch.
    """

    @staticmethod
    def create(ID, age):
        """
        :type ID: str
        :type age: int
        :rtype: GetUsersReqBody
        """

        return GetUsersReqBody(
            ID=ID,
            age=age,
        )

    def __init__(self, json=None, **kwargs):
        if json is None and not kwargs:
            raise ValueError('No data or kwargs present')

        class_name = 'GetUsersReqBody'
        create_error = '{cls}: unable to create {prop} from value: {val}: {err}'
        required_error = '{cls}: missing required property {prop}'

        data = json or kwargs

        property_name = 'ID'
        val = data.get(property_name)
        if val is not None:
            datatypes = [str]
            try:
                self.ID = client_support.val_factory(val, datatypes)
            except ValueError as err:
                raise ValueError(create_error.format(cls=class_name, prop=property_name, val=val, err=err))
        else:
            raise ValueError(required_error.format(cls=class_name, prop=property_name))

        property_name = 'age'
        val = data.get(property_name)
        if val is not None:
            datatypes = [int]
            try:
                self.age = client_support.val_factory(val, datatypes)
            except ValueError as err:
                raise ValueError(create_error.format(cls=class_name, prop=property_name, val=val, err=err))
        else:
            raise ValueError(required_error.format(cls=class_name, prop=property_name))

    def __str__(self):
        return self.as_json(indent=4)

    def as_json(self, indent=0):
        return client_support.to_json(self, indent=indent)

    def as_dict(self):
        return client_support.to_dict(self)
