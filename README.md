# pubsub
Simple publish subscribtion service to be used with GRPC streaming
# Description
Goal of this library is to be very simple, and only requirements is to broadcast events to all his subscribers.
Events must be sent in order.

# How to use
More docs will come but here short desc.

Implementation:
- Subscriber Interface needs to be implemented. You can check example folder.
- Eventer Interface needs to be implemented . There is already base Event implemented which you can extend.


Event has couple of fields
- Type
- Topic
- TopicId
- Application

For topic handling it is split to have topic and topic ID. When benchmarking I find this useful to avoid additional parsing. 

Application on event is basicaly subscriber ID. This is useful if you have subscribers that are basicaly application
consuming your app to do some work. So when your subscriber fires some request, you want all events that triggered 
by that request to be responded as part stream messaging.

# TODO 
- Removing subscribtion by queueing message. This is needed when subscriber fires request with application and wants all events to be fired to him.
Then you want subscribe -> fire events -> unsubscribe on subscriberID (Application) 
To avoid any process blocking unsubscribing must be queued as event.
- Add more documentation

