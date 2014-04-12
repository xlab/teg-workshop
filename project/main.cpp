#include <QApplication>
#include "qtquick2applicationviewer.h"

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);

    QtQuick2ApplicationViewer viewer;
    viewer.setMainQmlFile(QStringLiteral("qml/tegview.qml"));
    viewer.showExpanded();

    return app.exec();
}
