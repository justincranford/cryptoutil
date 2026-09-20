@post-checkout:
    uvx pre-commit run --hook-stage post-checkout

@pre-commit:
    uvx pre-commit run --hook-stage pre-commit

@pre-push:
    uvx pre-commit run --hook-stage pre-push
