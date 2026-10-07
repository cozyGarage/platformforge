def budget_minutes(slo, days=30):
    return (1 - slo) * days * 24 * 60


def burn_rate(errors, total, slo):
    if total == 0:
        return 0.0
    return (errors / total) / (1 - slo)


def should_page(long_window, short_window, slo, threshold=14.4):
    return all(burn_rate(e, t, slo) >= threshold for e, t in (long_window, short_window))


def hours_to_exhaustion(rate, days=30):
    return float("inf") if rate == 0 else days * 24 / rate
