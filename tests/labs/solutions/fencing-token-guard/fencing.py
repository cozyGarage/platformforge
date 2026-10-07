class StaleTokenError(Exception):
    pass


class LeaseStore:
    def __init__(self):
        self._token = 0
        self.holder = None

    def acquire(self, owner):
        self._token += 1
        self.holder = owner
        return self._token


class FencedStorage:
    def __init__(self):
        self._data = {}
        self._highest = 0
        self._seen = set()

    def write(self, key, value, token, request_id=None):
        if token < self._highest:
            raise StaleTokenError(f"token {token} < {self._highest}")
        if request_id is not None and request_id in self._seen:
            return False
        self._highest = token
        if request_id is not None:
            self._seen.add(request_id)
        self._data[key] = value
        return True

    def read(self, key):
        return self._data.get(key)
