"""Transform external Kafka payloads into internal command structures."""
from __future__ import annotations

from datetime import datetime, timedelta, timezone
from typing import Any, Dict, Iterable, Mapping

ISO_Z_SUFFIX = "Z"


class UnknownEventError(RuntimeError):
    """Raised when the incoming payload does not match any known schema."""


def map_external_event(payload: Mapping[str, Any]) -> Dict[str, Any]:
    """Convert a raw Kafka payload into RecordAnalyticsEventCommand data."""

    matcher = _EventMatcher()
    return matcher.match(payload)


class _EventMatcher:
    """Encapsulates heuristics for identifying external event shapes."""

    def __init__(self) -> None:
        pass

    def match(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        if self._is_iam_registered(payload):
            return self._handle_iam_registered(payload)
        if self._is_guides_challenges(payload):
            return self._handle_guides_challenges(payload)
        if self._is_execution_analytics(payload):
            return self._handle_execution_analytics(payload)
        if self._is_community_registration(payload):
            return self._handle_community_registration(payload)
        if self._is_community_profile_updated(payload):
            return self._handle_community_profile_updated(payload)
        if self._is_solution_completed(payload):
            return self._handle_solution_completed(payload)

        raise UnknownEventError("Unsupported analytics event payload signature")

    # --- IAM -----------------------------------------------------------------

    @staticmethod
    def _is_iam_registered(payload: Mapping[str, Any]) -> bool:
        return {
            "userId",
            "email",
            "firstName",
            "lastName",
            "provider",
        }.issubset(payload.keys()) and "registeredAt" in payload

    def _handle_iam_registered(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        occurred_at = _datetime_from_array(payload["registeredAt"])
        normalized_payload = dict(payload)
        normalized_payload["registeredAtIso"] = occurred_at.isoformat()

        return {
            "eventType": "iam.user-registered",
            "source": "iam-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }

    # --- Guides --------------------------------------------------------------

    @staticmethod
    def _is_guides_challenges(payload: Mapping[str, Any]) -> bool:
        return {"guideId", "challengeId", "occurredAt"}.issubset(payload.keys())

    def _handle_guides_challenges(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        occurred_at = _datetime_from_epoch(payload["occurredAt"])

        normalized_payload = dict(payload)
        normalized_payload["occurredAtIso"] = occurred_at.isoformat()

        return {
            "eventType": "guides.challenge-linked",
            "source": "guides-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }

    # --- Execution -----------------------------------------------------------

    @staticmethod
    def _is_execution_analytics(payload: Mapping[str, Any]) -> bool:
        return {"execution_id", "status", "timestamp"}.issubset(payload.keys())

    def _handle_execution_analytics(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        occurred_at = _datetime_from_iso(payload["timestamp"])

        normalized_payload = dict(payload)
        normalized_payload["timestampIso"] = occurred_at.isoformat()

        return {
            "eventType": "execution.analytics",
            "source": "execution-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }

    # --- Community -----------------------------------------------------------

    @staticmethod
    def _is_community_registration(payload: Mapping[str, Any]) -> bool:
        return {
            "userId",
            "profileId",
            "username",
            "occurredOn",
        }.issubset(payload.keys()) and payload.get("profileUrl") is None

    def _handle_community_registration(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        occurred_at = _datetime_from_array(payload["occurredOn"])

        normalized_payload = dict(payload)
        normalized_payload["occurredOnIso"] = occurred_at.isoformat()

        return {
            "eventType": "community.user-registered",
            "source": "community-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }

    @staticmethod
    def _is_community_profile_updated(payload: Mapping[str, Any]) -> bool:
        return {
            "userId",
            "profileId",
            "username",
            "occurredOn",
        }.issubset(payload.keys()) and payload.get("profileUrl") is not None

    def _handle_community_profile_updated(
        self, payload: Mapping[str, Any]
    ) -> Dict[str, Any]:
        occurred_at = _datetime_from_array(payload["occurredOn"])
        normalized_payload = dict(payload)
        normalized_payload["occurredOnIso"] = occurred_at.isoformat()

        return {
            "eventType": "community.profile-updated",
            "source": "community-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }

    # --- Challenge completion ------------------------------------------------

    @staticmethod
    def _is_solution_completed(payload: Mapping[str, Any]) -> bool:
        required_keys = {
            "studentId",
            "challengeId",
            "solutionId",
            "experiencePointsEarned",
            "occurredOn",
        }
        return required_keys.issubset(payload.keys())

    def _handle_solution_completed(self, payload: Mapping[str, Any]) -> Dict[str, Any]:
        occurred_at = _datetime_from_array(payload["occurredOn"])

        normalized_payload = dict(payload)
        normalized_payload["occurredOnIso"] = occurred_at.isoformat()

        return {
            "eventType": "challenges.solution-completed",
            "source": "challenges-service",
            "occurredAt": occurred_at.isoformat(),
            "payload": normalized_payload,
        }


def _datetime_from_array(values: Iterable[int]) -> datetime:
    seq = list(values)
    if len(seq) < 6:
        raise ValueError("Expected at least year..second values for timestamp array")

    year, month, day, hour, minute, second = seq[:6]
    nanosecond = seq[6] if len(seq) > 6 else 0
    dt = datetime(year, month, day, hour, minute, second, tzinfo=timezone.utc)
    microseconds = nanosecond // 1000
    return dt + timedelta(microseconds=microseconds)


def _datetime_from_epoch(value: Any) -> datetime:
    return datetime.fromtimestamp(float(value), tz=timezone.utc)


def _datetime_from_iso(value: str) -> datetime:
    working = value
    if working.endswith(ISO_Z_SUFFIX):
        working = working[:-1] + "+00:00"

    if "." in working:
        base, fraction_and_tz = working.split(".", 1)
        frac_digits = []
        tz_part = ""

        for char in fraction_and_tz:
            if char.isdigit():
                frac_digits.append(char)
            else:
                tz_part = fraction_and_tz[fraction_and_tz.index(char) :]
                break
        else:
            tz_part = ""

        fraction = "".join(frac_digits)
        fraction = (fraction + "000000")[:6]
        working = f"{base}.{fraction}{tz_part}"

    return datetime.fromisoformat(working)
