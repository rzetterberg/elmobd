OBDCommands use a req/res model. However, I'll like to add Events! In additions to a req/res model, Events implement an event driven model. Events are made up of a Trigger & Action

Whenever an Event.trigger.callback is positive, a list of Event.actions.callbacks functions are called.

An example, to set a speed limit of 100 mph, there's a need to continuously check speed's value for moments when the limit of 100 mph is exceeded. Then send an email + display a warning as a response to over-speeding.

To achieve this, add a watch Event on speed command, made up of a Trigger & Action.

There's an event loop constantly checking values of all OBDCommands being watched with appropriate delay.

Event.trigger.callback should return true when speed value > 100 mph.

All Event.actions.callback are called as a response to a positive Event.trigger.callback

Event.actions.callback will

send an email
display a warning.
Example usages is shown in example6

Inspired by: OBD python, https://python-obd.readthedocs.io/en/latest/Async%20Connections/

