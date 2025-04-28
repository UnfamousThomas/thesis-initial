# Service

The service provides a quality-of-life feature within the operator to simplify working with custom resources programmatically.

## Purpose

Creating custom resources in Kubernetes programmatically usually involves using the Kubernetes API directly. This often requires setting up ClusterRoles and other complex configurations. The goal of this service is to simplify the process by providing an easy-to-use API that handles the necessary configurations automatically.

## API Routes

The service currently provides the following API routes for managing resources:

| **Method**  | **Route**                          | **Description**                             |
|-------------|------------------------------------|---------------------------------------------|
| `POST`      | `/server`                          | Create a new server.                        |
| `DELETE`    | `/server`                          | Delete an existing server.                  |
| `POST`      | `/server/pod/labels`               | Add labels to the server's pod.             |
| `DELETE`    | `/server/pod/labels`               | Remove labels from the server's pod.        |
| `POST`      | `/fleet`                           | Create a new fleet.                         |
| `DELETE`    | `/fleet`                           | Remove an existing fleet.                   |
| `POST`      | `/scaler`                          | Create a new autoscaler.                    |
| `DELETE`    | `/scaler`                          | Delete an existing autoscaler.              |

Additional routes (e.g., `GET /server`) are planned for future improvements.

## JSON Structures

The JSON structures for most API calls are similar, with differences primarily in the specifications for each resource. Below are the JSON formats for creating and deleting servers and fleets.

### Create Server

The JSON for creating a new server is structured as follows:

```json
{
  "server": {
    "metadata": {
      "name": "server",
      "namespace": "default",
      "labels": {
        "label1": "value1"
      }
    },
    "spec": {
      "timeout": "5m",
      "allowForceDelete": false,
      "pod": {
        "containers": [
          {
            "name": "game-server",
            "image": "nginx:latest"
          }
        ]
      }
    }
  }
}
```

This JSON is similar to the manifest defined in the [Server](server.md) section.

### Create Fleet

A fleet's JSON structure follows this pattern:

```json
{
  "fleet": {
    "metadata": {
      "name": "fleet",
      "namespace": "default",
      "labels": {
        "label1": "value1"
      }
    },
    "spec": {
      "server": {
        "timeout": "5m",
        "allowForceDelete": false,
        "pod": {
          "containers": [
            {
              "name": "game-server",
              "image": "nginx:latest"
            }
          ]
        }
      },
      "scaling": {
        "replicas": 3,
        "prioritizeAllowed": true,
        "agePriority": "oldest_first"
      }
    }
  }
}
```

### Delete Resource

To delete a resource (e.g., server, fleet, scaler), the request should include the following JSON:

```json
{
  "metadata": {
    "name": "server",
    "namespace": "default"
  },
  "force": false
}
```

- The `force` field determines whether the resource should be deleted even if dependent resources exist. The default value is `false` and has no effect on GameAutoscalers.

### Add Pod Labels

To add or update labels for a server's pod:

```json
{
  "metadata": {
    "name": "server",
    "namespace": "default",
    "labels": {
      "label1": "val1",
      "label2": "val2"
    }
  }
}
```

- Note: This will overwrite any existing labels with the same key.

### Remove Pod Label

To remove a specific label from a server’s pod:

```json
{
  "metadata": {
    "name": "server",
    "namespace": "default"
  },
  "label": "label1"
}
```

This will remove the label with the key `label1`.

## Future Enhancements

In the future, we plan to add more routes, such as `GET /server`, to allow for more flexible management of game servers and related resources. Ideally, a dedicated API spec will also be setup for easier understanding of the documentation.
