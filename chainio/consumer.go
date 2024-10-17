package chainio

// BeatConsumer defines a supplementary component that should be used by
// subsystems which implement the `Consumer` interface. It partially implements
// the `Consumer` interface by providing the method `ProcessBlock` such that
// subsystems don't need to re-implement it.
//
// While inheritance is not commonly used in Go, subsystems embedding this
// struct cannot pass the interface check for `Consumer` because the `Name`
// method is not implemented, which gives us a "mortise and tenon" structure.
// In addition to reducing code duplication, this design allows `ProcessBlock`
// to work on the concrete type `Beat` to access its internal states.
type BeatConsumer struct {
	// BlockbeatChan is a channel to receive blocks from Blockbeat. The
	// received block contains the best known height and the txns confirmed
	// in this block.
	BlockbeatChan chan Blockbeat

	// name is the name of the consumer which embeds the BlockConsumer.
	name string

	// quit is a channel that closes when the BlockConsumer is shutting
	// down.
	//
	// NOTE: this quit channel should be mounted to the same quit channel
	// used by the subsystem.
	quit chan struct{}

	// currentBeat is the current beat of the consumer.
	currentBeat Blockbeat
}

// NewBeatConsumer creates a new BlockConsumer.
func NewBeatConsumer(quit chan struct{}, name string) BeatConsumer {
	b := BeatConsumer{
		BlockbeatChan: make(chan Blockbeat),
		quit:          quit,
		name:          name,
	}

	return b
}

// ProcessBlock takes a blockbeat and sends it to the blockbeat channel.
//
// NOTE: part of the `chainio.Consumer` interface.
func (b *BeatConsumer) ProcessBlock(beat Blockbeat) error {
	// Update the current height.
	beat.logger().Tracef("set current height for [%s]", b.name)
	b.currentBeat = beat

	select {
	// Send the beat to the blockbeat channel. It's expected that the
	// consumer will read from this channel and process the block. Once
	// processed, it should return the error or nil to the beat.Err chan.
	case b.BlockbeatChan <- beat:
		beat.logger().Tracef("Sent blockbeat to [%s]", b.name)

	case <-b.quit:
		beat.logger().Debugf("[%s] received shutdown before sending "+
			"beat", b.name)

		return nil
	}

	// Check the beat's err chan. We expect the consumer to call
	// `beat.NotifyBlockProcessed` to send the error back to the beat.
	select {
	case err := <-beat.errChan():
		beat.logger().Debugf("[%s] processed beat: err=%v", b.name, err)

		return err

	case <-b.quit:
		beat.logger().Debugf("[%s] received shutdown", b.name)
	}

	return nil
}
