# Kubernetes Controller & Custom Resource Requirements

## 📦 Custom Resources Overview

| CRD                   | Description                                                                 |
|-----------------------|-----------------------------------------------------------------------------|
| `Server`              | Wrapper around a Pod with sidecar, enforces delete protection via finalizer |
| `Fleet`               | A group of `Server` resources with scaling and smart deletion logic         |
| `GameType`            | Manages one or two Fleets and handles rolling updates on spec change        |
| `GameTypeAutoscaler`  | Scales `GameType` based on HTTP/webhook policy and sync schedule            |

---

## 🧠 Functional Requirements

### 1. `Server` Controller

#### Core Behavior
- Add a finalizer on all `Server` resources upon creation.
- Prevent deletion until the sidecar signals readiness to delete.

#### Pod Creating
- Create a pod matching the specifics in the spec
- Give the pod the same finalizer as the server has

#### Timeout Handling
- If `timeout` is specified:
    - Allow deletion after the timeout has passed since deletion started, even if the sidecar has not approved it.

---

### 2. `Fleet` Controller

#### Server Group Management
- Create `Server` resources using `FleetSpec`.

#### Replica Management
- Maintain number of `Server` resources equal to `FleetScaling.Replicas`.

#### Deletion Strategy
- If `prioritizeAllowed` is `true`, delete `Server` resources where deletion is allowed first.
- Apply `agePriority`:
    - `oldest_first`: delete the oldest eligible servers first.
    - `newest_first`: delete the newest eligible servers first.

#### Status Updates
- Set `.status.currentReplicas` to match active managed `Server` count.

---

### 3. `GameType` Controller

#### Fleet Lifecycle
- Create one `Fleet` per `GameType`, based on `fleetSpec`.
- Store active fleet's name in `.status`.

#### Rolling Update Strategy
- On `fleetSpec` change:
    - Create a new Fleet.
    - Wait for it to reach readiness.
    - Scale old Fleet to 0 and mark it for deletion.


---

### 4. `GameTypeAutoscaler` Controller

#### Target Selection
- Target the `GameType` specified by `GameName`.

#### Sync Triggers
- If `synctype` is `fixedinterval`, evaluate every `fixedInterval` seconds.

#### Policy Evaluation
- If `policytype` is `webhook`:
    - Send HTTP request to the defined `url` or Kubernetes `service` + `path`.
    - Parse response to determine the desired replica count.

#### Scaling Action
- Patch `GameType` replicas to match target from webhook.

---

## 🛡️ Non-Functional Requirements

### Observability
- Log all reconciliation steps with detailed reasoning using slog

### Resilience
- Gracefully handle missing or deleted dependent resources.

---

## 🔁 Controller Responsibility Matrix

| Controller              | Watches              | Creates/Manages                              |
|-------------------------|----------------------|----------------------------------------------|
| `Server` Controller     | `Server`             | Finalizers for Server and pod, timeout logic |
| `Fleet` Controller      | `Fleet`              | `Server` resources                           |
| `GameType` Controller   | `GameType`, `Fleet`  | Active Fleet lifecycle                       |
| `Autoscaler` Controller | `GameTypeAutoscaler` | Patches `GameType.spec.scaling`              |
