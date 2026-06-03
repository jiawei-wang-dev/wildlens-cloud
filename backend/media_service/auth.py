from __future__ import annotations

from fastapi import Header, HTTPException, status


def validate_cognito_token(token: str) -> str:
    """Return a fake owner id until AWS Cognito JWT verification is integrated."""
    # TODO: Validate AWS Cognito JWTs using env-provided region, user pool id,
    # and app client id before trusting the returned subject as owner_id.
    if not token:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing bearer token")
    return "fake-user-id"


def get_current_user_id(authorization: str | None = Header(default=None)) -> str:
    if not authorization:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing Authorization header")

    scheme, _, token = authorization.partition(" ")
    if scheme.lower() != "bearer" or not token:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Expected Authorization: Bearer <token>")

    return validate_cognito_token(token)
