# pubsub
Another Simple publish subscribtion service to be used with GRPC streaming
# Description
Goal of this library is to be very simple, and only requirements is to broadcast events to all his subscribers.
Events must be sent in order.

# How to use
Subscriber and Eventer must be implemented.

### Subscriber:
- GetId - method in order to keep pool list with ids
- Send(event) - callback function which is called when sending event.

### Eventer:
- GetTopic
- GetTopicId

For topic handling it is split to have topic and topic ID. I find this useful for handling ALL subscribtion easier.
Pool will check all subscribtions with:
- topic
- topic.topicId
    
USE "Topic" as domain and topicID as subdomain. 
When subscribing consider that you need format your topics **topic[.topicId]**
TopicID is not neccessary to be used.

## Changing subscribtion by queueing event
This is added so that you can handle dynamic resource subscribtions. 

Now this requires sub/unsub which can be done by queueing **SubEventUpdate**.
So you can do 
> SubEventUpdate, event1, event2, event3, SubEventUpdate

This helps so that you can keep order of things happening, and avoiding some additinal
checks is message sent.



