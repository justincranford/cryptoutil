@pre-commit:
    pre-commit run --hook-stage pre-commit

@post-checkout:
    pre-commit run --hook-stage post-checkout

@pre-push:
    pre-commit run --hook-stage pre-push
