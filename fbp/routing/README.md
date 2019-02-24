# routing

Nodes for routing messages. 

- Fanning
 -- Fan-In:
  -- `MergeN`: route N channels to 1
  -- `MergeTwo`: route 2 channels to 1
 -- Fan-Out:
  -- `SplitN`: broadcast messages from 1 channel to N channels
  -- `SplitTwo`: broadcast messages from 1 channel to 2 channels
- `Drain`: empties its input channel. You can use middleware-nodes as end-nodes this way.
- More to be implemented, suggestions welcome.
