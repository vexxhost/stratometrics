CREATE TABLE instance_events on cluster '{cluster}' AS instance_events_local
ENGINE = Distributed('{cluster}', default, instance_events_local, rand());
