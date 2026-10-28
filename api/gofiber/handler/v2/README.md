# Friend v2 + MQTT + Elastic/Backend Sequence Flow

## Sequence Diagram: Friend v2 Endpoints & MQTT Integration

```mermaid
sequenceDiagram
    participant Client as Game Client
    participant API as Fiber API Router
    participant Elasticsearch as Elasticsearch<br/>(Handler/Repository/Index)
    participant Redis as Redis Cache
    participant MQTT as MQTT<br/>(Handler/Service/Broker)
    participant Presence as PresenceService

    Note over Client,MQTT: === Friend v2 List Flow ===
    Client->>API: GET /friend/v2/list
    API->>Elasticsearch: OnlineFriendStatesWithElasticHandler()
    
    Elasticsearch->>Elasticsearch: Get player UUID from session/params
    Elasticsearch->>Elasticsearch: LoadFromES(playerUUID)
    Elasticsearch->>Elasticsearch: Query friend relationships
    
    alt No ES record found
        Elasticsearch->>Elasticsearch: LoadFromDBOrEmpty()
        Elasticsearch->>Elasticsearch: Save/Index document
    end
    
    Elasticsearch->>Elasticsearch: ToPlayerOnlineFriend()
    Elasticsearch->>Elasticsearch: LoadOnlineFriendStateFromCache()
    Elasticsearch->>Redis: MultiGet(online/{uuid}...)
    Redis-->>Elasticsearch: HeartbeatForm[] (online friends)
    Elasticsearch->>Elasticsearch: Build FriendStatesResponse
    Elasticsearch-->>API: JSON Response
    API-->>Client: 200 OK {onlineFriends, offlineFriends}

    Note over Client,MQTT: === Friend v2 Request Flow ===
    Client->>API: POST /friend/v2/request<br/>{action, target_uuid}
    API->>Elasticsearch: FriendRequestWithElasticHandler()
    
    Elasticsearch->>Elasticsearch: Validate input
    Elasticsearch->>Elasticsearch: CheckOrBuildFriendRelationshipExistInES()
    Elasticsearch->>Elasticsearch: Check/Get relationship doc
    
    Elasticsearch->>Elasticsearch: UpdateOneRelationship(action, targetUUID)
    Elasticsearch->>Elasticsearch: Update document<br/>(follow/unfollow/block/unblock)
    Elasticsearch-->>API: 201 Created
    API-->>Client: Success response

    Note over Client,MQTT: === Player Heartbeat → Presence Notification Flow ===
    Client->>API: POST /player/heartbeat<br/>{region, lobby, room, ...}
    API->>API: PlayerPostHeartbeat()
    
    API->>Redis: SETEX online/{uuid}<br/>(HeartbeatForm JSON, 15s)
    Redis-->>API: OK
    
    alt First time online (new cache entry)
        API->>Presence: PublishForFriendListCreation()
        Presence->>Elasticsearch: LoadFromES(playerUUID)
        Elasticsearch->>Elasticsearch: Query friend relationships
        Elasticsearch-->>Presence: FriendRelationship
        
        Presence->>Presence: CalculateFollowedByExcludeBlocking()
        loop For each friend
            Presence->>Redis: GET online/{friend_uuid}
            Redis-->>Presence: HeartbeatForm
            
            alt Version matches
                Presence->>MQTT: Publish({topic, content})
                MQTT-->>Client: MQTT message (friend online)
            end
        end
    end
    
    API-->>Client: 200 OK "Aliveness reported"

    Note over Client,MQTT: === MQTT Client Request Flow ===
    Client->>API: POST /mqtt/client/request<br/>{action, target, content}
    API->>MQTT: MqttPublishRequestHandler()
    
    MQTT->>MQTT: Validate input<br/>Build MqttMessage
    MQTT->>MQTT: PublishWithTimeout(5s, message)
    MQTT-->>Client: MQTT message delivered<br/>(game_inv, dir_cht, etc.)
    MQTT-->>API: Success
    API-->>Client: 200 OK Success response

    Note over Client,MQTT: === MQTT Broker Webhook Flow ===
    MQTT->>API: POST /mqtt/m5/authonregister<br/>{clientid, username, ...}
    API->>MQTT: MqttAuthOnRegisterM5Handler()
    MQTT->>MQTT: Validate client credentials
    MQTT-->>API: 200 OK {result: allow}
    API-->>MQTT: Auth result
    
    MQTT->>API: POST /mqtt/m5/authonsubscribe<br/>{clientid, topic, ...}
    API->>MQTT: MqttAuthOnSubscribeM5Handler()
    MQTT->>MQTT: Check subscription permissions
    MQTT-->>API: 200 OK {result: allow}
    API-->>MQTT: Auth result
    
    MQTT->>API: POST /mqtt/m5/authonpublish<br/>{clientid, topic, payload, ...}
    API->>MQTT: MqttAuthOnPublishM5Handler()
    MQTT->>MQTT: Check publish permissions
    MQTT-->>API: 200 OK {result: allow}
    API-->>MQTT: Auth result
```

## Key Flows Summary

### 1. **Friend List v2** (`GET /friend/v2/list`)
- Loads friend relationships from **Elasticsearch**
- Checks online status from **Redis cache** (`online/{uuid}`)
- Returns combined online/offline friend list

### 2. **Friend Request v2** (`POST /friend/v2/request`)
- Updates friend relationships directly in **Elasticsearch**
- Supports: follow, unfollow, block, unblock actions

### 3. **Player Heartbeat** (`POST /player/heartbeat`)
- Saves player state to **Redis** with 15s TTL
- Triggers **PresenceService** to notify friends via **MQTT**
- Friends receive real-time online notifications

### 4. **MQTT Client Request** (`POST /mqtt/client/request`)
- Client publishes game invitations, direct chat via HTTP API
- API forwards to **MQTT Broker** for delivery

### 5. **MQTT Broker Webhooks** (`/mqtt/m5/authon*`)
- **MQTT Broker** calls API for authentication/authorization
- API validates client permissions for register/subscribe/publish

## Data Flow: Online State Detection

```
Player Heartbeat → Redis (online/{uuid}) → Friend List Query → Redis MultiGet → Online Friends List
                                      ↓
                              MQTT Presence Notification → Friends' Clients
```

## Backend Components

- **Elasticsearch**: Stores friend relationships (follows, blocks)
- **Redis**: Caches player online state (`online/{uuid}` with HeartbeatForm JSON)
- **MQTT Broker**: Real-time messaging for presence notifications and game events
