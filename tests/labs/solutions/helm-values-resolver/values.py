import copy


def deep_merge(base, override):
    result = copy.deepcopy(base)
    for key, value in override.items():
        if isinstance(value, dict) and isinstance(result.get(key), dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = copy.deepcopy(value)
    return result


def _scalar(raw):
    if raw in ("true", "false"):
        return raw == "true"
    if raw.lstrip("-").isdigit():
        return int(raw)
    return raw


def set_flag(values, expr):
    key, _, raw = expr.partition("=")
    node = _scalar(raw)
    for part in reversed(key.split(".")):
        node = {part: node}
    return deep_merge(values, node)


def resolve(defaults, *files, sets=()):
    result = copy.deepcopy(defaults)
    for layer in files:
        result = deep_merge(result, layer)
    for expr in sets:
        result = set_flag(result, expr)
    return result
