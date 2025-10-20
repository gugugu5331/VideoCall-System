package queue

// Standardized task types (Redis List)
const (
    // User
    TaskUserRegister       = "user.register"
    TaskUserLogin          = "user.login"
    TaskUserProfileUpdate  = "user.profile_update"
    TaskUserStatusChange   = "user.status_change"

    // Meeting
    TaskMeetingCreate      = "meeting.create"
    TaskMeetingEnd         = "meeting.end"
    TaskMeetingRecording   = "meeting.recording_process"

    // Media
    TaskMediaTranscode     = "media.transcode"
    TaskMediaUploadToMinio = "media.upload_to_minio"

    // AI
    TaskAISpeechRecognition   = "ai.speech_recognition"
    TaskAIEmotionDetection    = "ai.emotion_detection"
    TaskAIDeepfakeDetection   = "ai.deepfake_detection"
)

// Pub/Sub channels and event types
const (
    ChannelUserEvents     = "user_events"
    ChannelMeetingEvents  = "meeting_events"
    ChannelMediaEvents    = "media_events"
    ChannelAIEvents       = "ai_events"
    ChannelSignalingEvents= "signaling_events"

    // Meeting events
    EventMeetingCreated   = "meeting.created"
    EventMeetingStarted   = "meeting.started"
    EventMeetingEnded     = "meeting.ended"
    EventUserJoined       = "meeting.user_joined"
    EventUserLeft         = "meeting.user_left"

    // Media events
    EventRecordingStarted = "recording.started"
    EventRecordingStopped = "recording.stopped"
    EventRecordingProcessed = "recording.processed"
    EventTranscodeCompleted = "transcode.completed"

    // AI events
    EventASRCompleted       = "speech_recognition.completed"
    EventEmotionCompleted   = "emotion_detection.completed"
    EventDeepfakeCompleted  = "deepfake_detection.completed"
)

