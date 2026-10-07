import math


def desired_replicas(current, metric, target, minimum=1, maximum=100, tolerance=0.1):
    ratio = metric / target
    if abs(ratio - 1) <= tolerance:
        return current
    return max(minimum, min(maximum, math.ceil(current * ratio)))


def stabilized(current, recommendations):
    latest = recommendations[-1]
    if latest > current:
        return latest
    return min(current, max(recommendations))
