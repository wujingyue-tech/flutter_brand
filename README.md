# flutter_brand

Go CLI that brands a Flutter iOS/Android template: Dart package name, home-screen display name, bundle id, splash, launcher icon.

It is not a Flutter feature. Install it, then run it from the **Flutter app root**:

```text
go install .
flutter_brand
flutter_brand name --pub my_app
flutter_brand display --title "My App" --title-zh "我的应用"
flutter_brand package --id com.example.myapp
flutter_brand splash --image path.png --image-dark path_dark.png
flutter_brand icon --image path.png
flutter_brand bump
```

`go run ../flutter_brand` from the Flutter git repo does not work (that repo has no `go.mod`). `$(go env GOPATH)/bin` must be on `PATH`.
